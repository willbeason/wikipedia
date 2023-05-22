package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/pages"
	"os"
	"sync"
)

// clean-wikipedia removes parts of articles we never want to analyze, such as xml tags, tables, and
// formatting directives.
func main() {
	err := mainCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.ExactArgs(1),
		Use:   `reference-network path/to/input`,
		Short: `Analyzes the network of references between biographical articles.`,
		RunE:  runCmd,
	}

	flags.Parallel(cmd)
	flags.IDs(cmd)

	return cmd
}

func runCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	parallel, err := cmd.Flags().GetInt(flags.ParallelKey)
	if err != nil {
		return err
	}

	inDB := args[0]

	source := pages.StreamDB(inDB, parallel)

	ctx, cancel := context.WithCancelCause(cmd.Context())

	docs, err := source(ctx, cancel)
	if err != nil {
		return err
	}

	result := make(map[string]uint32)
	resultMtx := sync.Mutex{}

	resultWork := jobs.Reduce(jobs.WorkBuffer, docs, func(page *documents.Page) error {
		resultMtx.Lock()
		result[page.Title] = page.Id
		resultMtx.Unlock()

		return nil
	})

	runner := jobs.NewRunner()

	resultWg := runner.Run(ctx, cancel, resultWork)
	resultWg.Wait()

	fmt.Println(len(result))

	return nil
}
