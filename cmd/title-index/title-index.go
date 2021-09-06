package main

import (
	"context"
	"os"
	"strings"

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
		Args:  cobra.ExactArgs(2),
		Use:   `view path/to/input id`,
		Short: `View a specific article by its identifier`,
		RunE:  runCmd,
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
	out := args[1]

	source := pages.StreamDB(inDB, parallel)

	errs, errsWg := jobs.Errors()

	ps, err := source(ctx, errs)

	results := makeIndex(ps)

	index := <-results

	err = protos.Write(out, index)
	if err != nil {
		errs <- err
	}

	close(errs)
	errsWg.Wait()

	return nil
}

func makeIndex(pages <-chan *documents.Page) <-chan *documents.TitleIndex {
	results := make(chan *documents.TitleIndex)

	go func() {
		result := &documents.TitleIndex{
			Titles: make(map[string]uint32),
		}

		for page := range pages {
			title := strings.ToLower(page.Title)
			result.Titles[title] = page.Id
		}

		results <- result
		close(results)
	}()

	return results
}
