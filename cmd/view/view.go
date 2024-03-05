package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"os"

	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/pages"
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
		Args:  cobra.ExactArgs(1),
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

	// We don't want to print pages in parallel.
	parallel := 1

	pageIDs, err := cmd.Flags().GetUintSlice(flags.IDsKey)
	if err != nil {
		return err
	}

	titles, err := cmd.Flags().GetStringSlice(flags.TitlesKey)
	if err != nil {
		return err
	}
	if len(titles) > 0 {

	}

	if len(pageIDs) == 0 && len(titles) == 0 {
		return fmt.Errorf("must specify at least one ID or title")
	}

	inDB := args[0]

	if _, err = os.Stat(inDB); err != nil {
		return fmt.Errorf("unable to open %q: %w", inDB, err)
	}

	var source func(ctx context.Context, cancel context.CancelCauseFunc) (<-chan *documents.Page, error)
	if len(pageIDs) == 0 {
		source = pages.StreamDB(inDB, parallel)
	} else {
		fmt.Println("Page IDs", pageIDs)
		source = pages.StreamDBKeys(inDB, parallel, pageIDs)
	}

	runner := jobs.NewRunner()

	ps, err := source(ctx, cancel)
	printWork := jobs.ForEach(jobs.WorkBuffer, ps, func(from *documents.Page) error {
		return nil
	})

	printWg := runner.Run(ctx, cancel, printWork)
	printWg.Wait()

	return nil
}
