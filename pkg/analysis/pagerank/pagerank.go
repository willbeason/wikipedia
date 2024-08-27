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
	"math"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
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
	weights := make(map[uint32]float64, len(links))
	for i := range links {
		weights[i] = 1.0 / float64(len(links))
	}
	fmt.Println(len(weights), "Weights")

	converged := atomic.Bool{}
	for i := range maxIterations {
		if converged.Load() {
			fmt.Printf("Converged early at iteration %d\n", i)
			break
		}

		nextWeights := iteratePageRank(ctx, errs, links, weights)

		go func(j int, before, after map[uint32]float64) {
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
	for k, weight := range weights {
		pageRanks[i] = PageRank{
			id:   k,
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

func iteratePageRank(ctx context.Context, errs chan<- error, network map[uint32][]uint32, weights map[uint32]float64) map[uint32]float64 {
	kvSource := jobs.NewSource(jobs.MapSourceFn(network))
	kvWg, kvJob, kvs := kvSource()
	go kvJob(ctx, errs)

	dampWeightMtx := sync.Mutex{}
	dampWeight := (1.0 - damping) / float64(len(network))

	rowMap := jobs.NewMap(jobs.ReduceToMany(jobs.MakeMap[uint32, float64],
		func(kv jobs.KV[uint32, []uint32], out map[uint32]float64) error {
			if len(kv.Value) == 0 {
				// This article links to nowhere, so all weight goes to damping.
				dampWeightMtx.Lock()
				dampWeight += damping * weights[kv.Key] / float64(len(network))
				dampWeightMtx.Unlock()
				return nil
			}

			linkToWeight := damping * weights[kv.Key] / float64(len(kv.Value))
			for _, to := range kv.Value {
				out[to] += linkToWeight
			}
			return nil
		},
	))
	rowMapWg, rowMapJob, weightsChan := rowMap(kvs)
	for range 64 {
		go rowMapJob(ctx, errs)
	}

	rowReduce := jobs.NewMap(jobs.ReduceToOne(jobs.MakeMap[uint32, float64], func(from, to map[uint32]float64) error {
		for k, v := range from {
			to[k] += v
		}
		return nil
	}))
	rowReduceWg, rowReduceJob, rowReduceFuture := rowReduce(weightsChan)
	go rowReduceJob(ctx, errs)

	kvWg.Wait()
	rowMapWg.Wait()
	rowReduceWg.Wait()

	nextWeights := <-rowReduceFuture

	// Add random traversal weight.
	for id := range network {
		nextWeights[id] += dampWeight
	}

	return nextWeights
}

type PageRank struct {
	id   uint32
	rank float64
}
