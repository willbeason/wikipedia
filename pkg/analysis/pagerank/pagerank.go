package pagerank

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/config"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"math"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const (
	damping           = 0.85
	convergeThreshold = 1e-8

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
			fmt.Printf("Converged early at iteration %d\n", i)
			break
		}

		nextWeights := iteratePageRank(ctx, errs, indexLinks, weights)

		go func(j int, before, after []float64) {
			diffs := 0.0
			for id := range before {
				diffs += math.Abs(before[id] - after[id])
			}

			fmt.Printf("%d: %.010f\n", i, diffs)

			if diffs < convergeThreshold {
				fmt.Printf("Converged early at iteration %d\n", j)
				converged.Store(true)
			}
		}(i, weights, nextWeights)

		weights = nextWeights
	}

	pageRanks := make([]PageRank, len(weights))
	i := 0
	for index, weight := range weights {
		pageRanks[i] = PageRank{
			id:   indexToId[index],
			rank: weight,
		}
		i++
	}
	fmt.Println(len(pageRanks), "PageRanks")

	sort.Slice(pageRanks, func(i, j int) bool {
		return pageRanks[i].rank > pageRanks[j].rank
	})

	titleIndex := <-titleIndexFuture

	for i := range 10 {
		title := titleIndex[pageRanks[i].id]
		rank := pageRanks[i].rank
		fmt.Printf("%d: %s, %.08f\n", i, title, rank)
	}

	//articleGenders := <-articleGendersFuture

	close(errs)
	errsWg.Wait()
	if ctx.Err() != nil {
		return context.Cause(ctx)
	}

	return nil
}

func makeIdDictionary(network map[uint32][]uint32) ([]uint32, map[uint32]int) {
	indexToId := make([]uint32, len(network))
	idToIndex := make(map[uint32]int)


	i := 0
	for k := range network {
		indexToId[i] = k
		idToIndex[k] = i
		i++
	}

	return indexToId, idToIndex
}

func convertToIndex(network map[uint32][]uint32, idToIndex map[uint32]int) [][]int {
	result := make([][]int, len(network))

	for k, v := range network {
		converted := make([]int, len(v))
		for i, id := range v {
			converted[i] = idToIndex[id]
		}
		result[idToIndex[k]] = converted
	}

	return result
}

func iteratePageRank(ctx context.Context, errs chan<- error, network [][]int, weights []float64) []float64 {
	start := time.Now()

	dampWeightMtx := sync.Mutex{}
	dampWeight := (1.0 - damping) / float64(len(network))

	// Initialize all possible keys
	nextWeights := make([]float64, len(weights))

	invNetworkSize := 1.0 / float64(len(network))
	for k, v := range network {
		if len(v) == 0 {
			// This article links to nowhere, so all weight goes to damping.
			dampWeightMtx.Lock()
			dampWeight += damping * weights[k] * invNetworkSize
			dampWeightMtx.Unlock()
			continue
		}

		linkToWeight := damping * weights[k] / float64(len(v))
		for _, to := range v {
			nextWeights[to] += linkToWeight
		}
	}


	// Add random traversal weight.
	for k := range nextWeights {
		nextWeights[k] += dampWeight
	}
	fmt.Println(time.Since(start))

	return nextWeights
}

type PageRank struct {
	id   uint32
	rank float64
}
