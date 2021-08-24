package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/willbeason/extract-wikipedia/pkg/db"

	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/flags"
	"github.com/willbeason/extract-wikipedia/pkg/jobs"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"
)

// clean-wikipedia removes parts of articles we never want to analyze, such as xml tags, tables, and
// formatting directives.

const (
	idsKey = "ids"
)

func main() {
	ctx := context.Background()

	err := mainCmd().ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.RangeArgs(1, 2),
		Use:  `clean-wikipedia path/to/input path/to/output`,
		Short: `Cleans an extracted set of Wikipedia articles by removing irrelevant pages and formatting
directives.`,
		RunE: runCmd,
	}

	flags.Parallel(cmd)

	cmd.Flags().UintSlice(idsKey, nil, "A list of specific article ids to check.")

	return cmd
}

func runCmd(cmd *cobra.Command, args []string) error {
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

	cleaned := make(chan db.MessageID, jobs.WorkBuffer)
	cleanWg := jobs.RunPage(cmd.Context(), parallel, cleanPages(cleaned), work, errs)
	go func() {
		cleanWg.Wait()
		close(cleaned)
	}()

	var writeWg *sync.WaitGroup
	if outDBPath == "" {
		writeWg = printPages(cleaned)
	} else {
		outDB := db.NewRunner(outDBPath, parallel)
		writeWg, err = outDB.Write(cleaned, errs)
		if err != nil {
			return err
		}
	}

	writeWg.Wait()
	close(errs)

	errsWg.Wait()

	return nil
}

func cleanPages(cleaned chan<- db.MessageID) jobs.Page {
	return func(page *documents.Page) error {
		page.Text = nlp.CleanArticle(page.Text)
		cleaned <- page

		return nil
	}
}

func printPages(cleaned <-chan db.MessageID) *sync.WaitGroup {
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
	result := make(chan uint32, jobs.WorkBuffer)

	go func() {
		for _, id := range ids {
			result <- uint32(id)
		}

		close(result)
	}()

	return result
}
