package pagerank

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/analysis/gender"
	"github.com/willbeason/wikipedia/pkg/config"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"path/filepath"
	"sort"
	"sync"
)

func RunComparePageRank(cmd *cobra.Command, cfg *config.ComparePageRank, corpusNames ...string) error {
	if len(corpusNames) != 2 {
		return fmt.Errorf("%w: must have exactly two corpora but got %+v", ErrPageRank, corpusNames)
	}
	beforeCorpus := corpusNames[0]
	afterCorpus := corpusNames[1]

	ctx, cancel := context.WithCancelCause(cmd.Context())

	errs := make(chan error, 1)
	errsWg := sync.WaitGroup{}
	errsWg.Add(1)
	go func() {
		for err := range errs {
			cancel(err)
			fmt.Println(err)
		}
		errsWg.Done()
	}()

	workspace, err := flags.GetWorkspacePath(cmd)
	if err != nil {
		return err
	}

	beforePageRankFile := filepath.Join(workspace, beforeCorpus, cfg.PageRanks)
	afterPageRankFile := filepath.Join(workspace, afterCorpus, cfg.PageRanks)

	beforePageRanksFuture := documents.ReadPageRanks(ctx, errs, beforePageRankFile)
	afterPageRanksFuture := documents.ReadPageRanks(ctx, errs, afterPageRankFile)

	beforePageRanks := <- beforePageRanksFuture
	afterPageRanks := <- afterPageRanksFuture

	beforeGenderFile := filepath.Join(workspace, beforeCorpus, cfg.GenderIndex)
	afterGenderFile := filepath.Join(workspace, afterCorpus, cfg.GenderIndex)

	beforeGenderIndexFuture := documents.ReadGenderMap(ctx, beforeGenderFile, errs)
	afterGenderIndexFuture := documents.ReadGenderMap(ctx, afterGenderFile, errs)

	beforeGenderIndex := <- beforeGenderIndexFuture
	afterGenderIndex := <- afterGenderIndexFuture

	beforeGenderPageRank := make(map[string]float64)
	afterGenderPageRank := make(map[string]float64)

	for id, pageRank := range beforePageRanks {
		pageGender, ok := beforeGenderIndex[id]
		if !ok {
			// Does not have gender entry so is not a human.
			continue
		}
		beforeGenderPageRank[pageGender] += pageRank
	}
	for id, pageRank := range afterPageRanks {
		pageGender, ok := afterGenderIndex[id]
		if !ok {
			// Does not have gender entry so is not a human.
			continue
		}
		afterGenderPageRank[pageGender] += pageRank
	}

	var results []GenderPageRank
	for pageGender := range beforeGenderPageRank {
		results = append(results, GenderPageRank{
			Gender: pageGender,
			Before: beforeGenderPageRank[pageGender],
			After:  afterGenderPageRank[pageGender],
		})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Before > results[j].Before
	})

	fmt.Println("Gender,Before,After")
	for _, result := range results {
		readable := gender.ReadableGender[result.Gender]
		if readable == "" {
			readable = result.Gender
		}
		fmt.Printf("%s,%g,%g\n", readable, result.Before, result.After)
	}

	close(errs)
	errsWg.Wait()
	if ctx.Err() != nil {
		return context.Cause(ctx)
	}

	return nil
}

type GenderPageRank struct {
	Gender string
	Before float64
	After float64
}
