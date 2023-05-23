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
	"regexp"
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

	idMap := make(map[string]uint32)
	titleMap := make(map[uint32]string)
	resultMtx := sync.Mutex{}

	docs, err := source(ctx, cancel)
	if err != nil {
		return err
	}

	idMapWork := jobs.Reduce(jobs.WorkBuffer, docs, func(page *documents.Page) error {
		resultMtx.Lock()
		idMap[page.Title] = page.Id
		titleMap[page.Id] = page.Title
		resultMtx.Unlock()

		return nil
	})

	runner := jobs.NewRunner()
	idMapWg := runner.Run(ctx, cancel, idMapWork)
	idMapWg.Wait()

	network := make(map[uint32][]uint32)
	networkMtx := sync.Mutex{}

	source2 := pages.StreamDB(inDB, parallel)
	docs2, err := source2(ctx, cancel)
	if err != nil {
		return err
	}

	networkWork := jobs.Reduce(jobs.WorkBuffer, docs2, func(page *documents.Page) error {
		from, foundFrom := idMap[page.Title]
		if !foundFrom {
			return fmt.Errorf("did not add ID for %q", page.Title)
		}

		matches := linkRegex.FindAllString(page.Text, -1)
		var tos []uint32
		for _, match := range matches {
			to, foundTo := idMap[match]
			if !foundTo {
				return fmt.Errorf("did not add ID for %q", page.Title)
			}

			tos = append(tos, to)
		}

		networkMtx.Lock()
		network[from] = tos
		networkMtx.Unlock()

		return nil
	})

	networkWg := runner.Run(ctx, cancel, networkWork)
	networkWg.Wait()

	fmt.Println(len(idMap))
	fmt.Println("Nodes:", len(network))

	totalEdges := 0
	singletons := 0

	for _, edges := range network {
		totalEdges += len(edges)
		if len(edges) == 0 {
			singletons++
		}
	}
	fmt.Println("Edges", totalEdges)
	fmt.Println("Singletons", singletons)

	return ctx.Err()
}

var linkRegex = regexp.MustCompile(`\[\[[^]]+]]`)
