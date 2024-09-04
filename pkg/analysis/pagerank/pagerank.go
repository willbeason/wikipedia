package pagerank

import (
	"context"
	"errors"
	"fmt"
	"github.com/willbeason/wikipedia/pkg/analysis/gender"
	"math"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/config"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/protos"
)

const (
	defaultDamping           = 0.85
	defaultConvergeThreshold = 1e-9

	defaultMaxIterations = 100
)

var ErrPageRank = errors.New("calculating PageRank")

func RunPageRank(cmd *cobra.Command, cfg *config.PageRank, corpusNames ...string) error {
	if len(corpusNames) != 1 {
		return fmt.Errorf("%w: must have exactly one corpus but got %+v", ErrPageRank, corpusNames)
	}
	corpusName := corpusNames[0]

	ctx, cancel := context.WithCancelCause(cmd.Context())

	errs := make(chan error, 1)
	errsWg := sync.WaitGroup{}
	errsWg.Add(1)
	go func() {
		for err := range errs {
			cancel(err)
		}
		errsWg.Done()
	}()

	workspace, err := flags.GetWorkspacePath(cmd)
	if err != nil {
		return err
	}

	titleIndexFile := filepath.Join(workspace, corpusName, cfg.Index)
	titleIndexFuture := documents.ReadReverseTitleMap(ctx, titleIndexFile, errs)

	articleGendersFuture := make(<-chan map[uint32]string)
	if cfg.GenderIndex != "" {
		fmt.Println("Started loading gender index.")
		genderIndexFile := filepath.Join(workspace, corpusName, cfg.GenderIndex)
		articleGendersFuture = documents.ReadGenderMap(ctx, genderIndexFile, errs)
	}

	linksFile := filepath.Join(workspace, corpusName, cfg.Links)
	linksFuture := documents.ReadLinksMap(ctx, linksFile, errs)

	fmt.Println("Started loading links.")
	links := <-linksFuture
	fmt.Println("Finished loading links.")
	fmt.Println(len(links), "Links")
	indexToId, idToIndex := makeIdDictionary(links)
	indexLinks := convertToIndex(links, idToIndex)

	var genderIndex map[uint32]string
	var filter []bool
	if cfg.GenderIndex != "" {
		fmt.Println("Finished loading gender index.")
		genderIndex = <-articleGendersFuture

		filter = make([]bool, len(links))
		for i := range filter {
			filter[i] = genderIndex[indexToId[i]] == cfg.GenderFilter
		}
	}

	calculator := Calculator{
		damping:       defaultDamping,
		threshold:     defaultConvergeThreshold,
		maxIterations: defaultMaxIterations,
		filter:        filter,
	}

	weights := calculator.calculatePageRank(indexLinks)
	fmt.Println(len(weights), "Weights")

	pageRanks := make([]*documents.PageRank, len(weights))
	i := 0
	for index, weight := range weights {
		pageRanks[i] = &documents.PageRank{
			Id:       indexToId[index],
			Pagerank: weight,
		}
		i++
	}
	fmt.Println(len(pageRanks), "PageRanks")

	sort.Slice(pageRanks, func(i, j int) bool {
		return pageRanks[i].Pagerank > pageRanks[j].Pagerank
	})

	titleIndex := <-titleIndexFuture

	i, n := 0, 0
	for n < 100 {
		if cfg.GenderIndex != "" {
			for genderIndex[pageRanks[i].Id] == "" {
				i++
			}
		}

		pageRank := pageRanks[i]
		title := titleIndex[pageRank.Id]
		rank := pageRank.Pagerank

		if cfg.GenderIndex == "" {
			fmt.Printf("%d,%q,%.020f\n", i, title, rank)
		} else {
			articleGender := gender.ReadableGender[genderIndex[pageRank.Id]]
			fmt.Printf("%d,%q,%s,%.020f\n", i, title, articleGender, rank)
		}

		i++
		n++
	}

	pageRanksSource := jobs.NewSource(jobs.SliceSourceFn(pageRanks))
	pageRanksSourceWg, pageRanksSourceJob, pageRanksChan := pageRanksSource()
	go pageRanksSourceJob(ctx, errs)

	outFile := filepath.Join(workspace, corpusName, cfg.Out)
	pageRanksSink := jobs.NewSink(protos.WriteFile[*documents.PageRank](outFile))
	pageRanksSinkWg, pageRanksSinkJob := pageRanksSink(pageRanksChan)
	go pageRanksSinkJob(ctx, errs)

	pageRanksSourceWg.Wait()
	pageRanksSinkWg.Wait()

	close(errs)
	errsWg.Wait()
	if ctx.Err() != nil {
		return context.Cause(ctx)
	}

	return nil
}

func makeIdDictionary(network map[uint32][]uint32) ([]uint32, map[uint32]uint32) {
	indexToId := make([]uint32, len(network))
	idToIndex := make(map[uint32]uint32)

	i := uint32(0)
	for k := range network {
		indexToId[i] = k
		idToIndex[k] = i
		i++
	}

	return indexToId, idToIndex
}

func convertToIndex(network map[uint32][]uint32, idToIndex map[uint32]uint32) [][]uint32 {
	result := make([][]uint32, len(network))

	for k, v := range network {
		converted := make([]uint32, len(v))
		for i, id := range v {
			converted[i] = idToIndex[id]
		}
		result[idToIndex[k]] = converted
	}

	return result
}

type Calculator struct {
	// Damping is the proportion of traffic which navigates via a link on the
	// current article. Each iteration, all weight is "damped" by multiplying by
	// this constant.
	damping float64

	// Threshold is the desired "accuracy" of the PageRank calculation.
	// Is the proportion of total PageRank beyond which we don't care about error.
	threshold     float64
	maxIterations int

	// filter tracks the set of articles which act as primary sources for damped weight.
	// If unset, all pages are considered primary sources.
	// If set, secondary sources still get dead-end weight to preserve network behavior.
	// (This is what lets us add subnetworks together.)
	filter []bool
}

// calculatePageRank calculates the unnormalized Pagerank for all nodes in a network.
//
//		network is a sparse array of links. Each element is a list of links from the node
//	 whose identifier corresponds to the index in the list.
//
// Uses an average PageRank of 1.0 instead of 1.0 / len(network) as n
func (c Calculator) calculatePageRank(network [][]uint32) []float64 {
	weights := make([]float64, len(network))
	if c.filter == nil {
		for i := range weights {
			weights[i] = 1.0
		}
	} else {
		for i := range weights {
			if c.filter[i] {
				weights[i] = 1.0
			}
		}
	}

	converged := atomic.Bool{}
	threshold := c.threshold * float64(len(network))
	for i := range c.maxIterations {
		if converged.Load() {
			break
		}

		nextWeights := c.iteratePageRank(network, weights)

		go func(j int, before, after []float64) {
			diffs := 0.0
			for id := range before {
				diffs += math.Abs(before[id] - after[id])
			}

			fmt.Printf("%d: %g\n", i, diffs)

			if diffs < threshold {
				fmt.Printf("Converged early at iteration %d\n", j)
				converged.Store(true)
			}
		}(i, weights, nextWeights)

		weights = nextWeights
	}

	return weights
}

func (c Calculator) iteratePageRank(network [][]uint32, weights []float64) []float64 {
	// Initialize all possible keys.
	// This is faster than reusing the same slice and resetting the values each time.
	nextWeights := make([]float64, len(weights))

	// dampWeight is the weight added to all articles every iteration.
	dampWeight := 1.0 - c.damping

	// deadEndWeight tracks the weight of all articles which link nowhere.
	deadEndWeight := 0.0
	for k, v := range network {
		if len(v) == 0 {
			// This article links to nowhere, so all weight is distributed evenly.
			deadEndWeight += weights[k]
		} else {
			// Evenly distribute this article's weight to all linked articles.
			linkToWeight := weights[k] / float64(len(v))
			for _, to := range v {
				nextWeights[to] += linkToWeight
			}
		}
	}

	// Damp deadEndWeight since some of that weight only goes to filtered pages.
	deadEndWeight *= c.damping
	// Since deadEndWeight tracks the total network weight to be evenly
	// distributed, we divide it by the network's size.
	deadEndWeight /= float64(len(network))

	// Add random traversal weight.
	if c.filter == nil {
		for k := range nextWeights {
			nextWeights[k] = nextWeights[k]*c.damping + dampWeight + deadEndWeight
		}
	} else {
		for k := range nextWeights {
			// All pages are damped and get the dead end weight.
			nextWeights[k] = nextWeights[k]*c.damping + deadEndWeight
			if c.filter[k] {
				// Only filtered articles get the damp weight.
				// Effectively this makes these articles the "source" of PageRank.
				nextWeights[k] += dampWeight
			}
		}
	}

	return nextWeights
}
