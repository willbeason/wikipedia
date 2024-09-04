package documents

import (
	"context"

	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/protos"
)

func ReadPageRanks(ctx context.Context, errs chan<- error, filename string) <-chan map[uint32]float64 {
	pageRankSource := jobs.NewSource(protos.ReadFile[PageRank](filename))
	_, sourceJob, pageRanks := pageRankSource()
	go sourceJob(ctx, errs)

	pageRankReduce := jobs.NewMap(jobs.ReduceToOne(jobs.MakeMap[uint32, float64], mergePageRankMap))
	_, reduceJob, futurePageRanks := pageRankReduce(pageRanks)
	go reduceJob(ctx, errs)

	return futurePageRanks
}

func mergePageRankMap(pageRank *PageRank, pageRankMap map[uint32]float64) error {
	pageRankMap[pageRank.Id] = pageRank.Pagerank
	return nil
}
