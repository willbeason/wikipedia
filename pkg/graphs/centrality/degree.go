package centrality

import (
	"strings"

	"github.com/willbeason/wikipedia/pkg/graphs"
)

func InDegree(id uint32, graph *graphs.Directed) int {
	result := 0

	for j, edges := range graph.Nodes {
		if id == j {
			// Ignore self-loops.
			continue
		}

		if edges[id] {
			result++
		}
	}

	return result
}

func OutDegree(id uint32, graph *graphs.Directed) int {
	return len(graph.Nodes[id])
}

func Normalize(title string) string {
	title = strings.TrimPrefix(title, "category:")
	//title = strings.TrimSuffix(title, ", california")
	//title = strings.TrimSuffix(title, " (california)")
	title = strings.Title(title)
	title = strings.ReplaceAll(title, " Of", " of")
	title = strings.ReplaceAll(title, " In", " in")
	title = strings.ReplaceAll(title, " The", " the")

	return title
}
