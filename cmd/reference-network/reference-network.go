package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/nlp"
	"github.com/willbeason/wikipedia/pkg/pages"
	"os"
	"regexp"
	"strings"
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
	female := make(map[uint32]bool)
	male := make(map[uint32]bool)
	resultMtx := sync.Mutex{}

	docs, err := source(ctx, cancel)
	if err != nil {
		return err
	}

	idMapWork := jobs.Reduce(jobs.WorkBuffer, docs, func(page *documents.Page) error {
		gender := nlp.DetermineGender(page.Text)

		resultMtx.Lock()
		idMap[page.Title] = page.Id
		titleMap[page.Id] = page.Title
		switch gender {
		case nlp.Female:
			female[page.Id] = true
		case nlp.Male:
			male[page.Id] = true
		}
		resultMtx.Unlock()

		return nil
	})

	runner := jobs.NewRunner()
	idMapWg := runner.Run(ctx, cancel, idMapWork)
	// Must fully wait for ID Map to be created and the Badger database closed before opening another connection.
	idMapWg.Wait()

	fmt.Println("Biographies:", len(idMap))
	fmt.Println("Women:", len(female))
	fmt.Println("Men:", len(male))

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
			// Strip square brackets.
			match = match[2 : len(match)-2]

			// Only consider before vertical bar.
			if idx := strings.Index(match, "|"); idx != -1 {
				match = match[:idx]
			}

			to, foundTo := idMap[match]
			if !foundTo {
				continue
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

	fmt.Println("Nodes:", len(network))

	totalEdges := 0
	f2f := 0
	f2m := 0
	m2f := 0
	m2m := 0
	singletons := 0

	for id, edges := range network {
		totalEdges += len(edges)

		for _, edge := range edges {
			switch {
			case female[id] && female[edge]:
				f2f++
			case female[id] && male[edge]:
				f2m++
			case male[id] && female[edge]:
				m2f++
			case male[id] && male[edge]:
				m2m++
			}
		}

		if len(edges) == 0 {
			singletons++
		}
	}

	fmt.Println("Edges", totalEdges)
	fmt.Println("Singletons", singletons)
	fmt.Println("Avg F2F", float64(f2f)/float64(len(female)))
	fmt.Println("Avg F2M", float64(f2m)/float64(len(female)))
	fmt.Println("Avg M2F", float64(m2f)/float64(len(male)))
	fmt.Println("Avg M2M", float64(m2m)/float64(len(male)))

	if ctx.Err() != nil {
		return context.Cause(ctx)
	}

	return nil
}

var linkRegex = regexp.MustCompile(`\[\[[^]]+]]`)
