package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/pages"
	"github.com/willbeason/wikipedia/pkg/protos"
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
		Args: cobra.ExactArgs(3),
		RunE: runCmd,
	}

	flags.Parallel(cmd)

	return cmd
}

func runCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	ctx := cmd.Context()

	parallel, err := cmd.Flags().GetInt(flags.ParallelKey)
	if err != nil {
		return err
	}

	inDB := args[0]
	inTitleIndex := args[1]
	outPageCategories := args[2]

	titleIndex := &documents.TitleIndex{}
	err = protos.Read(inTitleIndex, titleIndex)
	if err != nil {
		return err
	}

	source := pages.StreamDB(inDB, parallel)

	errs, errsWg := jobs.Errors()

	ps, err := source(ctx, errs)

	results := makePageCategories(titleIndex, ps)

	pageCategories := <-results

	err = protos.Write(outPageCategories, pageCategories)
	if err != nil {
		errs <- err
	}

	close(errs)
	errsWg.Wait()

	fmt.Println(documents.Missed)

	return nil
}

func makePageCategories(titleIndex *documents.TitleIndex, pages <-chan *documents.Page) <-chan *documents.PageCategories {
	results := make(chan *documents.PageCategories)

	go func() {
		categorizer := &documents.Categorizer{
			TitleIndex: titleIndex,
		}

		result := &documents.PageCategories{
			Pages: make(map[uint32]*documents.Categories),
		}

		for page := range pages {
			result.Pages[page.Id] = categorizer.Categorize(page)
		}

		results <- result
	}()

	return results
}

type freq struct {
	c int
	s string
}
