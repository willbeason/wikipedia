package main

import (
	"context"
	"fmt"
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
			inDBPath := args[0]
			var outDBPath string
			if len(args) > 1 {
				outDBPath = args[1]
			}

			parallelJobs, err := cmd.Flags().GetInt(flags.ParallelKey)
			if err != nil {
				return err
			}

			pageIds, err := cmd.Flags().GetUintSlice(idsKey)
			if err != nil {
				return err
			}

			inDB, err := badger.Open(badger.DefaultOptions(inDBPath))
			defer func() {
				_ = inDB.Close()
			}()
			if err != nil {
				return err
			}

			errs, errsWg := jobs.Errors()
			var work <-chan proto.Message

			if len(pageIds) == 0 {
				work = jobs.Walk(cmd.Context(), inDB, newPage, parallelJobs, errs)
			} else {
				work = jobs.IDs(inDB, newPage, pageIds, errs)
			}

			cleaned := make(chan jobs.MessageID, 100)

			cleanWg := jobs.RunProto(parallelJobs, cleanDocuments(cleaned), work, errs)

			var outDB *badger.DB
			if outDBPath != "" {
				outDB, err = badger.Open(badger.DefaultOptions(outDBPath))
				if err != nil {
					return err
				}
			}

			var writeWg *sync.WaitGroup
			if outDB == nil {
				writeWg = printPages(cleaned)
			} else {
				writeWg = jobs.WriteProtos(outDB, parallelJobs, cleaned, errs)
			}

			cleanWg.Wait()
			close(cleaned)

			writeWg.Wait()
			close(errs)

			errsWg.Wait()

			err = inDB.Close()
			if err != nil {
				return err
			}

			err = outDB.Close()
			if err != nil {
				return err
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
