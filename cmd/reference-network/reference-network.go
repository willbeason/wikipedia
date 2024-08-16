package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/nlp"
	"github.com/willbeason/wikipedia/pkg/pages"
)

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

	parallel, err := flags.GetParallel(cmd)
	if err != nil {
		return err
	}

	inDB := args[0]

	source := pages.StreamDB[documents.Page](inDB, parallel)

	ctx, cancel := context.WithCancelCause(cmd.Context())

	idMap := make(map[string]uint32)
	titleMap := make(map[uint32]string)
	female := make(map[uint32]bool)
	male := make(map[uint32]bool)
	other := make(map[uint32]bool)
	unknown := make(map[uint32]bool)
	resultMtx := sync.Mutex{}

	docs, err := source(ctx, cancel)
	if err != nil {
		return err
	}

	checker, err := documents.NewInfoboxChecker(documents.PersonInfoboxes())
	if err != nil {
		return err
	}

	idMapWork := jobs.Reduce(jobs.WorkBuffer, docs, func(page *documents.Page) error {
		if !checker.Matches(page.Text) {
			// Not a biography.
			resultMtx.Lock()
			idMap[page.Title] = page.Id
			titleMap[page.Id] = page.Title
			resultMtx.Unlock()
			return nil
		}

		gender := nlp.InferGender(page.Text)

		resultMtx.Lock()
		idMap[page.Title] = page.Id
		titleMap[page.Id] = page.Title
		switch gender {
		case nlp.Female:
			female[page.Id] = true
		case nlp.Male:
			male[page.Id] = true
		case nlp.Nonbinary, nlp.Multiple:
			other[page.Id] = true
		case nlp.Unknown:
			unknown[page.Id] = true
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
	fmt.Println("Unknown:", len(unknown))

	network := make(map[uint32][]uint32)
	networkMtx := sync.Mutex{}

	source2 := pages.StreamDB[documents.Page](inDB, parallel)
	docs2, err := source2(ctx, cancel)
	if err != nil {
		return err
	}

	networkWork := jobs.Reduce(
		jobs.WorkBuffer,
		docs2,
		addPageToNetwork(idMap, &networkMtx, network),
	)

	networkWg := runner.Run(ctx, cancel, networkWork)
	networkWg.Wait()

	if ctx.Err() != nil {
		return context.Cause(ctx)
	}

	fmt.Println("Nodes:", len(network))

	totalEdges := 0

	matrix := make(map[FromTo]uint64)

	singletons := 0

	for id, edges := range network {
		totalEdges += len(edges)

		for _, edge := range edges {
			from := NonBiography
			switch {
			case female[id]:
				from = Female
			case male[id]:
				from = Male
			case unknown[id]:
				from = Unknown
			}

			to := NonBiography
			switch {
			case female[edge]:
				to = Female
			case male[edge]:
				to = Male
			case unknown[edge]:
				to = Unknown
			}

			ft := FromTo{From: from, To: to}

			matrix[ft]++
		}

		if len(edges) == 0 {
			singletons++
		}
	}

	order := []Class{Female, Male, Unknown, NonBiography}

	fmt.Println("Edges", totalEdges)
	fmt.Println("Singletons", singletons)

	columnSize := make([]uint64, len(order))

	for row := range order {
		for column := range order {
			ft := FromTo{From: order[column], To: order[row]}
			columnSize[column] += matrix[ft]
		}
	}

	for row := range order {
		for column := range order {
			ft := FromTo{From: order[column], To: order[row]}
			fmt.Printf("%.05f,", float64(matrix[ft])/float64(columnSize[column]))
		}
		fmt.Printf("\n")
	}

	return nil
}

func addPageToNetwork(
	idMap map[string]uint32,
	networkMtx *sync.Mutex,
	network map[uint32][]uint32,
) func(page *documents.Page) error {
	return func(page *documents.Page) error {
		from, foundFrom := idMap[page.Title]
		if !foundFrom {
			return fmt.Errorf("%w: did not add ID for %q", jobs.ErrStream, page.Title)
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
	}
}

var linkRegex = regexp.MustCompile(`\[\[[^]]+]]`)

type Class string

const (
	Female       Class = "female"
	Male         Class = "male"
	Unknown      Class = "unknown"
	NonBiography Class = "non-biography"
)

type FromTo struct {
	From Class
	To   Class
}
