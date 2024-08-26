package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	gender2 "github.com/willbeason/wikipedia/pkg/analysis/gender"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/pages"
	"github.com/willbeason/wikipedia/pkg/protos"
)

// clean removes parts of articles we never want to analyze, such as xml tags, tables, and
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

	genders, err := protos.ReadOne[documents.GenderIndex]("")
	if err != nil {
		return err
	}

	idMapWork := jobs.Reduce(ctx, jobs.WorkBuffer, docs, func(page *documents.Page) error {
		gender := genders.Genders[page.Id]

		resultMtx.Lock()
		idMap[page.Title] = page.Id
		titleMap[page.Id] = page.Title
		switch gender {
		case gender2.WomanGender:
			female[page.Id] = true
		case gender2.ManGender:
			male[page.Id] = true
		case gender2.NonBinaryGender, gender2.ConflictingClaims:
			other[page.Id] = true
		default:
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

	network := make(map[uint32][]uint32)
	networkMtx := sync.Mutex{}

	source2 := pages.StreamDB[documents.Page](inDB, parallel)
	docs2, err := source2(ctx, cancel)
	if err != nil {
		return err
	}

	networkWork := jobs.Reduce(ctx, jobs.WorkBuffer, docs2, addPageToNetwork(idMap, &networkMtx, network))

	networkWg := runner.Run(ctx, cancel, networkWork)
	networkWg.Wait()

	fmt.Println("Nodes:", len(network))

	// Initialize the network so every node has equal weight.
	weights := make(map[uint32]float64, len(network))
	startWeight := 1.0 / float64(len(network))
	for id := range network {
		weights[id] = startWeight
	}

	fw, mw := RelativeWeights(female, male, weights)
	fmt.Printf("Start (%.03f, %.03f)\n", fw, mw)

	for i := range 40 {
		nextWeights := make(map[uint32]float64, len(network))
		dampWeight := (1.0 - damping) / float64(len(network))

		for id, tos := range network {
			var linkToWeight float64

			if len(tos) > 0 {
				linkToWeight = damping * weights[id] / float64(len(tos))
			} else {
				// This article links to nowhere, so all weight goes to damping.
				dampWeight += damping * weights[id] / float64(len(network))
			}

			// Add weight from links.
			for _, to := range tos {
				nextWeights[to] += linkToWeight
			}
		}

		// Add random traversal weight.
		for id := range network {
			nextWeights[id] += dampWeight
		}

		diffs := 0.0
		for id, before := range weights {
			diffs += math.Abs(before - nextWeights[id])
		}
		fmt.Printf("Distance %d %.03f\n", i, diffs)

		weights = nextWeights
		fwi, mwi := RelativeWeights(female, male, weights)
		fmt.Printf("Iteration %d (%.03f, %.03f)\n", i, fwi, mwi)
	}

	if ctx.Err() != nil {
		return context.Cause(ctx)
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

		// Force-exclude self reference.
		tos := []uint32{}
		seen := map[uint32]bool{from: true}

		for _, match := range matches {
			// Strip square brackets.
			match = match[2 : len(match)-2]

			// Only consider before vertical bar.
			if idx := strings.Index(match, "|"); idx != -1 {
				match = match[:idx]
			}

			// Ignore references to non-biographies.
			to, foundTo := idMap[match]
			if !foundTo {
				continue
			}

			// Don't add duplicates.
			if seen[to] {
				continue
			}
			seen[to] = true

			tos = append(tos, to)
		}

		networkMtx.Lock()
		network[from] = tos
		networkMtx.Unlock()

		return nil
	}
}

const damping = 0.99

var linkRegex = regexp.MustCompile(`\[\[[^]]+]]`)

func RelativeWeights(female, male map[uint32]bool, weights map[uint32]float64) (float64, float64) {
	femaleWeight := 0.0
	maleWeight := 0.0

	for id, weight := range weights {
		if female[id] {
			femaleWeight += weight
		} else if male[id] {
			maleWeight += weight
		}
	}

	return femaleWeight, maleWeight
}
