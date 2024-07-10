package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/charts"
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

	source := pages.StreamDB(inDB, parallel)

	ctx, cancel := context.WithCancelCause(cmd.Context())

	idMap := make(map[string]uint32)
	titleMap := make(map[uint32]string)

	biographies := make(map[uint32]bool)
	female := make(map[uint32]bool)
	male := make(map[uint32]bool)
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
			resultMtx.Lock()
			idMap[page.Title] = page.Id
			titleMap[page.Id] = page.Title
			resultMtx.Unlock()
			return nil
		}

		gender := nlp.DetermineGender(page.Text)

		resultMtx.Lock()
		idMap[page.Title] = page.Id
		titleMap[page.Id] = page.Title

		biographies[page.Id] = true
		switch gender {
		case nlp.Female:
			female[page.Id] = true
		case nlp.Male:
			male[page.Id] = true
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

	fmt.Println("Articles:", len(idMap))
	fmt.Println("Biographies:", len(female)+len(male)+len(unknown))
	fmt.Println("Women:", len(female))
	fmt.Println("Men:", len(male))
	fmt.Println("Unknown", len(unknown))

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

		// Force-exclude self reference.
		var tos []uint32
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
	})

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

	for i := range 100 {
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
		if diffs < 1e-8 {
			fmt.Printf("Converged early at iteration %d\n", i)

			beforePageRanks := make([]PageRank, len(weights))
			{
				j := 0
				for id, rank := range weights {
					beforePageRanks[j] = PageRank{id: id, rank: rank}
					j++
				}
			}

			afterPageRanks := make([]PageRank, len(nextWeights))
			{
				j := 0
				for id, rank := range nextWeights {
					afterPageRanks[j] = PageRank{id: id, rank: rank}
					j++
				}
			}

			sort.Slice(beforePageRanks, func(i, j int) bool {
				return beforePageRanks[i].rank > beforePageRanks[j].rank
			})
			sort.Slice(afterPageRanks, func(i, j int) bool {
				return afterPageRanks[i].rank > afterPageRanks[j].rank
			})

			for n, pr := range beforePageRanks {
				if afterPageRanks[n].id != pr.id {
					fmt.Println("First Difference at index", n)
					break
				}
			}

			weights = nextWeights

			break
		}
		fmt.Printf("%d: %.010f\n", i, diffs)

		weights = nextWeights
	}

	pageRanks := make([]PageRank, len(weights))
	{
		i := 0
		for id, rank := range weights {
			pageRanks[i] = PageRank{id: id, rank: rank}
			i++
		}
	}

	sort.Slice(pageRanks, func(i, j int) bool {
		return pageRanks[i].rank > pageRanks[j].rank
	})

	bins := charts.LogarithmicBins(10, 18, math.Pow(10.0, 1.0/3.0))

	femaleBins := make([]int, len(bins)+1)
	maleBins := make([]int, len(bins)+1)

	femaleTraffic := 0.0
	maleTraffic := 0.0

	// Only track biographies.
	n := 0
	for _, pageRank := range pageRanks {
		if !biographies[pageRank.id] {
			continue
		}

		bin := toBin(n, bins)
		if female[pageRank.id] {
			femaleBins[bin]++
			femaleTraffic += pageRank.rank
		} else if male[pageRank.id] {
			maleBins[bin]++
			maleTraffic += pageRank.rank
		}

		n++
	}

	fmt.Printf("Female Rank: %.08f", femaleTraffic)
	fmt.Printf("Male Rank:   %.08f", maleTraffic)
	fmt.Printf("Disparity:   %.08f", maleTraffic/femaleTraffic)

	for i := range len(bins) {
		fmt.Printf("%d,%d,%d\n", i, femaleBins[i], maleBins[i])
	}

	if ctx.Err() != nil {
		return context.Cause(ctx)
	}

	return nil
}

func toBin(n int, bins []int) int {
	for i, bin := range bins {
		if n <= bin {
			return i
		}
	}
	return len(bins)
}

const damping = 0.85

var linkRegex = regexp.MustCompile(`\[\[[^]]+]]`)

type PageRank struct {
	id   uint32
	rank float64
}

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
