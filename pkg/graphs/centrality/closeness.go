package centrality

import "github.com/willbeason/wikipedia/pkg/graphs"

type ClosenessHarmonicResult struct {
	Closeness, Harmonic float64
}

func ClosenessHarmonic(id uint32, graph *graphs.Directed) ClosenessHarmonicResult {
	closeness := 0.0
	harmonic := 0.0

	curDistance := 1
	visited := make(map[uint32]bool, len(graph.Nodes))
	visited[id] = true

	curLayer := graph.Nodes[id]
	nextLayer := make(map[uint32]bool)

	for len(curLayer) > 0 {
		for n := range curLayer {
			if visited[n] {
				continue
			}

			visited[n] = true

			for child := range graph.Nodes[n] {
				if !visited[child] {
					nextLayer[child] = true
				}
			}

			closeness += float64(curDistance)
			harmonic += 1 / float64(curDistance)
		}

		curDistance++

		curLayer = nextLayer
		nextLayer = make(map[uint32]bool)
	}

	closeness = 1.0 / closeness
	closeness *= float64(len(graph.Nodes) - 1)

	harmonic /= float64(len(graph.Nodes) - 1)

	return ClosenessHarmonicResult{Closeness: closeness, Harmonic: harmonic}
}
