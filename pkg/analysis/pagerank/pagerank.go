package pagerank

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/config"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/protos"
	"math"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
)

const (
	damping           = 0.85
	convergeThreshold = 1e-10

	maxIterations = 100
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

	//genderIndexFile := filepath.Join(workspace, corpusName, cfg.GenderIndex)
	//articleGendersFuture := documents.ReadGenderMap(ctx, genderIndexFile, errs)

	linksFile := filepath.Join(workspace, corpusName, cfg.Links)
	linksFuture := documents.ReadLinksMap(ctx, linksFile, errs)

	fmt.Println("Started loading links.")
	links := <-linksFuture
	fmt.Println("Finished loading links.")
	fmt.Println(len(links), "Links")
	// Initialize weights equally.
	weights := make([]float64, len(links))
	invLenLinks := 1.0 / float64(len(links))
	for i := range weights {
		weights[i] = invLenLinks
	}
	fmt.Println(len(weights), "Weights")

	indexToId, idToIndex := makeIdDictionary(links)
	indexLinks := convertToIndex(links, idToIndex)

	converged := atomic.Bool{}
	for i := range maxIterations {
		if converged.Load() {
			break
		}

		nextWeights := iteratePageRank(indexLinks, weights)

		go func(j int, before, after []float64) {
			diffs := 0.0
			for id := range before {
				diffs += math.Abs(before[id] - after[id])
			}

			fmt.Printf("%d: %g\n", i, diffs)

			if diffs < convergeThreshold {
				fmt.Printf("Converged early at iteration %d\n", j)
				converged.Store(true)
			}
		}(i, weights, nextWeights)

		weights = nextWeights
	}

	pageRanks := make([]*documents.PageRank, len(weights))
	i := 0
	for index, weight := range weights {
		pageRanks[i] = &documents.PageRank{
			Id:   indexToId[index],
			Pagerank: weight,
		}
		i++
	}
	fmt.Println(len(pageRanks), "PageRanks")

	sort.Slice(pageRanks, func(i, j int) bool {
		return pageRanks[i].Pagerank > pageRanks[j].Pagerank
	})

	titleIndex := <-titleIndexFuture

	for i := range 10 {
		title := titleIndex[pageRanks[i].Id]
		rank := pageRanks[i].Pagerank
		fmt.Printf("%d: %s, %.020f\n", i, title, rank)
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

func iteratePageRank(network [][]uint32, weights []float64) []float64 {
	// Initialize all possible keys.
	// This is faster than reusing the same slice and resetting the values each time.
	nextWeights := make([]float64, len(weights))

	invNetworkSize := 1.0 / float64(len(network))
	dampWeight := (1.0 - damping) * invNetworkSize
	deadEndWeight := damping * invNetworkSize
	for k, v := range network {
		if len(v) == 0 {
			// This article links to nowhere, so all weight goes to damping.
			dampWeight += weights[k] * deadEndWeight
		} else {
			linkToWeight := weights[k] / float64(len(v))
			for _, to := range v {
				nextWeights[to] += linkToWeight
			}
		}
	}

	// Add random traversal weight.
	for k, v := range nextWeights {
		nextWeights[k] = v * damping + dampWeight
	}

	return nextWeights
}
