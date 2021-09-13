package centrality

import "github.com/willbeason/wikipedia/pkg/graphs"

func ClosenessHarmonic(id uint32, graph *graphs.Directed, cache *graphs.ShortestCache) (closeness, harmonic float64) {
	closeness = 0.0
	harmonic = 0.0

	for j := range graph.Nodes {
		if id == j {
			continue
		}

		dist := graphs.FindDistance(id, j, *graph, cache)
		if dist == 0 {
			continue
		}

		closeness += float64(dist)
		harmonic += 1.0 / float64(dist)
	}

	closeness = 1.0 / closeness
	closeness *= float64(len(graph.Nodes) - 1)

	harmonic /= float64(len(graph.Nodes) - 1)

	return closeness, harmonic
}
