package centrality

import (
	"strings"

	"github.com/willbeason/wikipedia/pkg/graphs"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

var caser = cases.Title(language.English)

func Normalize(title string) string {
	title = strings.TrimPrefix(title, "category:")
	// title = strings.TrimSuffix(title, ", california")
	// title = strings.TrimSuffix(title, " (california)")
	title = caser.String(title)
	title = strings.ReplaceAll(title, " Of", " of")
	title = strings.ReplaceAll(title, " In", " in")
	title = strings.ReplaceAll(title, " The", " the")

	return title
}
