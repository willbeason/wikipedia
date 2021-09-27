package centrality

import (
	"fmt"
	"math"

	"github.com/willbeason/wikipedia/pkg/graphs"
)

func Markhov(g *graphs.Directed, threshold float64, maxIters int) map[uint32]float64 {
	next := make(map[uint32]float64, len(g.Nodes))

	for n := range g.Nodes {
		next[n] = 1.0
	}

	var current map[uint32]float64

	dist := math.Sqrt(float64(len(g.Nodes)))
	iters := 0

	for dist >= threshold {
		if iters > maxIters {
			fmt.Println("Max Iterations Reached")
			break
		}

		current = next
		next = make(map[uint32]float64, len(g.Nodes))

		for from, tos := range g.Nodes {
			toAdd := current[from] / float64(len(tos))

			for to := range tos {
				next[to] += toAdd
			}
		}

		dist = 0.0

		for n := range g.Nodes {
			diff := current[n] - next[n]
			dist += diff * diff
		}

		dist = math.Sqrt(dist)
		iters++

		fmt.Println(iters, dist)
	}

	fmt.Println("Converged in", iters, "iterations")

	return next
}
