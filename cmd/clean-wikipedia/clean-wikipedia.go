package main

import (
	"context"
	"fmt"
	"github.com/willbeason/extract-wikipedia/pkg/db"
	"os"
	"sync"

	"github.com/dgraph-io/badger/v3"
	"github.com/willbeason/extract-wikipedia/pkg/flags"
	"github.com/willbeason/extract-wikipedia/pkg/jobs"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"
	"google.golang.org/protobuf/proto"

	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/documents"
)

// clean-wikipedia removes parts of articles we never want to analyze, such as xml tags, tables, and
// formatting directives.

const (
	idsKey = "ids"
)

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.MinimumNArgs(1),
		Use:  `clean-wikipedia path/to/input path/to/output`,
		Short: `Cleans an extracted set of Wikipedia articles by removing irrelevant pages and formatting
directives.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			var outDBPath string
			if len(args) > 1 {
				outDBPath = args[1]
			}

			parallel, err := cmd.Flags().GetInt(flags.ParallelKey)
			if err != nil {
				return err
			}

			pageIDs, err := cmd.Flags().GetUintSlice(idsKey)
			if err != nil {
				return err
			}

			inDBPath := args[0]
			inDB := db.NewRunner(inDBPath, parallel)

			errs, errsWg := jobs.Errors()
			work := make(chan *documents.Page)

			var inWg *sync.WaitGroup
			if len(pageIDs) == 0 {
				inWg, err = inDB.Process(cmd.Context(), documents.ReadPages(work), errs)
				if err != nil {
					return err
				}
			} else {
				inWg, err = inDB.ProcessIDs(cmd.Context(), documents.ReadPages(work), toUint32Chan(pageIDs), errs)
				if err != nil {
					return err
				}
			}
			go func() {
				inWg.Wait()
				close(work)
			}()

			cleaned := make(chan jobs.MessageID, 100)

			cleanWg := jobs.RunProto(parallel, cleanDocuments(cleaned), work, errs)

			var outDB *badger.DB
			if outDBPath != "" {
				outDB, err = badger.Open(badger.DefaultOptions(outDBPath))
				if err != nil {
					return err
				}

				defer func() {
					_ = outDB.Close()
				}()
			}

			var writeWg *sync.WaitGroup
			if outDB == nil {
				writeWg = printPages(cleaned)
			} else {
				writeWg = jobs.WriteProtos(outDB, parallel, cleaned, errs)
			}

			cleanWg.Wait()
			close(cleaned)

			writeWg.Wait()
			close(errs)

			errsWg.Wait()

			if outDB != nil {
				err = outDB.Close()
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	flags.Parallel(cmd)

	cmd.Flags().UintSlice(idsKey, nil, "A list of specific article ids to check.")

	return cmd
}

func main() {
	ctx := context.Background()

	err := mainCmd().ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}

func newPage() proto.Message {
	return &documents.Page{}
}

func cleanDocuments(cleaned chan<- jobs.MessageID) jobs.Proto {
	return func(p proto.Message) error {
		page, ok := p.(*documents.Page)
		if !ok {
			return fmt.Errorf("got message type %T, want %T", p, &documents.Page{})
		}

		page.Text = nlp.CleanArticle(page.Text)
		cleaned <- page

		return nil
	}
}

func printPages(cleaned <-chan jobs.MessageID) *sync.WaitGroup {
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for p := range cleaned {
			page, ok := p.(*documents.Page)
			if !ok {
				panic(fmt.Errorf("got message type %T, want %T", p, &documents.Page{}))
			}

			fmt.Println(page.Text)
		}

		wg.Done()
	}()

	return &wg
}

func toUint32Chan(ids []uint) chan uint32 {
	result := make(chan uint32, 100)

	go func() {
		for _, id := range ids {
			result <- uint32(id)
		}

		close(result)
	}()

	return result
}
