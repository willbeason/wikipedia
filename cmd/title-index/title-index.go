package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/environment"
	"github.com/willbeason/wikipedia/pkg/flags"
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
		Use:   `view path/to/input`,
		Short: `View specific articles by identifier (--ids) or title (--titles)`,
		RunE:  runCmd,
	}

	flags.Parallel(cmd)
	flags.IDs(cmd)
	flags.Titles(cmd)

	return cmd
}

func runCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	ctx, cancel := context.WithCancelCause(cmd.Context())

	parallel, err := cmd.Flags().GetInt(flags.ParallelKey)
	if err != nil {
		return err
	}

	inDB := ""
	if len(args) > 0 {
		inDB = args[0]
	} else {
		inDB = filepath.Join(environment.WikiPath, "extracted.db")
	}

	out := ""
	if len(args) > 1 {
		out = args[1]
	} else {
		out = filepath.Join(environment.WikiPath, environment.TitleIndex)
	}

	source := pages.StreamDB(inDB, parallel)

	ps, err := source(ctx, cancel)
	if err != nil {
		return err
	}

	results := makeIndex(ps)

	index := <-results

	err = protos.Write(out, index)
	if err != nil {
		return err
	}

	return nil
}

func makeIndex(pages <-chan *documents.Page) <-chan *documents.TitleIndex {
	results := make(chan *documents.TitleIndex)

	go func() {
		result := &documents.TitleIndex{
			Titles: make(map[string]uint32),
		}

		for page := range pages {
			result.Titles[page.Title] = page.Id
		}

		results <- result
		close(results)
	}()

	return results
}
