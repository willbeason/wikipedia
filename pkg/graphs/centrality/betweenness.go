package centrality

import (
	"github.com/willbeason/wikipedia/pkg/graphs"
	"github.com/willbeason/wikipedia/pkg/jobs"
)

func Betweenness(g *graphs.Directed) map[uint32]float64 {
	result := make(map[uint32]float64, len(g.Nodes))

	work := make(chan uint32, jobs.WorkBuffer)

	go func() {
		for id := range g.Nodes {
			work <- id
		}
	}()

	return result
}

func shortestPaths(g *graphs.Directed) <-chan []uint32 {
	result := make(chan []uint32, jobs.WorkBuffer)

	go func() {
	}()

	return result
}
