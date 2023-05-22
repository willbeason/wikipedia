package main

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/pages"
	"github.com/willbeason/wikipedia/pkg/protos"
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

	return pages.Run(cmd.Context(), source, parallel, idName(), protos.NoSink)
}

// namesToIDs
func namesToIDs(ctx context.Context, parallel int, source pages.Source) (map[string]uint32, error) {
	result := make(map[string]uint32)

	err := pages.Run(ctx, source, parallel, idName(), idNameMapSink(result))

	if err != nil {
		return nil, err
	}

	return result, nil
}

func idNameMapSink(out map[string]uint32) protos.Sink {
	return func(ctx context.Context, ids <-chan protos.ID, errors chan<- error) (*sync.WaitGroup, error) {
		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			for id := range ids {
				page := id.(*documents.Page)
				out[page.Title] = page.Id
			}

			wg.Done()
		}()

		return &wg, nil
	}
}

func idName() func(chan<- protos.ID) jobs.Page {
	return func(ids chan<- protos.ID) jobs.Page {
		return func(page *documents.Page) error {
			// We don't need the page text for this step.
			page.Text = ""
			return nil
		}
	}
}
