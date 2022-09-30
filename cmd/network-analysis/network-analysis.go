package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/graphs"
	"github.com/willbeason/wikipedia/pkg/graphs/centrality"
	"github.com/willbeason/wikipedia/pkg/protos"
)

func main() {
	ctx := context.Background()

	err := mainCmd().ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(4),
		RunE: runCmd,
	}

	flags.Parallel(cmd)

	return cmd
}

func runCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	inCategories := args[0]
	inTitles := args[1]
	inNodes := args[2]
	// outCentrality := args[3]

	pageTitles := &documents.TitleIndex{}

	err := protos.Read(inTitles, pageTitles)
	if err != nil {
		return err
	}

	reverseTitles := make(map[uint32]string)
	for title, id := range pageTitles.Titles {
		reverseTitles[id] = title
	}

	pageCategories := &documents.PageCategories{}

	err = protos.Read(inCategories, pageCategories)
	if err != nil {
		return err
	}

	nodesBytes, err := os.ReadFile(inNodes)
	if err != nil {
		return err
	}

	nodes := strings.Split(string(nodesBytes), "\n")
	ids := toIDs(nodes, pageTitles.Titles)

	graph := &graphs.Directed{
		Nodes: make(map[uint32]map[uint32]bool, len(ids)),
	}

	for id := range ids {
		categories := pageCategories.Pages[id].Categories
		nodeEdges := make(map[uint32]bool, len(categories))

		for _, category := range categories {
			if !ids[category] {
				// Ignore edges to nodes not in this subgraph.
				continue
			}

			nodeEdges[category] = true
		}

		graph.Nodes[id] = nodeEdges
	}

	runMarkov(graph, reverseTitles)

	return nil
}

func toIDs(categories []string, titles map[string]uint32) map[uint32]bool {
	result := make(map[uint32]bool, len(categories))

	for _, category := range categories {
		id, found := titles[category]
		if !found {
			panic(fmt.Sprintf("did not find page %q", category))
		}

		result[id] = true
	}

	return result
}

func runMarkov(g *graphs.Directed, reverseTitles map[uint32]string) {
	weights := centrality.Markov(g, 1e-10, 1000)

	nodes := make([]uint32, len(g.Nodes))
	idx := 0

	for n := range g.Nodes {
		nodes[idx] = n
		idx++
	}

	sort.Slice(nodes, func(i, j int) bool {
		return weights[nodes[i]] > weights[nodes[j]]
	})

	for _, n := range nodes[:200] {
		fmt.Println(n, reverseTitles[n], weights[n])
	}
}
