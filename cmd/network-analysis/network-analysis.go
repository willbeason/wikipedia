package main

import (
	"context"
	"fmt"
	"github.com/willbeason/wikipedia/pkg/graphs/centrality"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/graphs"
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

	parallel, err := cmd.Flags().GetInt(flags.ParallelKey)
	if err != nil {
		return err
	}

	inCategories := args[0]
	inTitles := args[1]
	inNodes := args[2]
	outCentrality := args[3]

	pageTitles := &documents.TitleIndex{}

	err = protos.Read(inTitles, pageTitles)
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

	nodesBytes, err := ioutil.ReadFile(inNodes)
	if err != nil {
		return err
	}

	nodes := strings.Split(string(nodesBytes), "\n")
	ids := toIDs(nodes, pageTitles.Titles)

	graph := &graphs.Directed{
		Nodes: make(map[uint32]map[uint32]bool, len(ids)),
	}

	ignored := &documents.PageCategories{}

	for id := range ids {
		categories := pageCategories.Pages[id].Categories
		nodeEdges := make(map[uint32]bool, len(categories))

		for _, category := range categories {
			if !ids[category] {
				// Ignore edges to nodes not in this subgraph.
				continue
			}

			//title1 := centrality.Normalize(reverseTitles[id])
			//title2 := centrality.Normalize(reverseTitles[category])
			//fmt.Printf("  %q -+ %q,\n", title1, title2)

			nodeEdges[category] = true
		}

		graph.Nodes[id] = nodeEdges
	}

	err = protos.Write("data/ignored-relationships-000.json", ignored)
	if err != nil {
		return err
	}

	//edgeRows := make(map[FromTo]RowEdge)
	//
	//lookID := uint32(0)
	//for id := range graph.Nodes {
	//	if lookID != 0 && id != lookID {
	//		continue
	//	}
	//
	//	cycle := graphs.FindPath(id, id, *graph)
	//	if len(cycle) == 0 {
	//		fmt.Printf("No cycle for %s\n", reverseTitles[id])
	//		continue
	//	} else {
	//		if id == lookID {
	//			for _, n := range cycle {
	//				fmt.Println(reverseTitles[n])
	//			}
	//		}
	//	}
	//
	//	var (
	//		from uint32
	//		to   uint32
	//	)
	//
	//	for i := 0; i < len(cycle); i++ {
	//		from = cycle[i]
	//
	//		if i == len(cycle)-1 {
	//			to = cycle[0]
	//		} else {
	//			to = cycle[i+1]
	//		}
	//
	//		ft := FromTo{From: from, To: to}
	//		er := edgeRows[ft]
	//		er.FromTo = ft
	//		er.FromTitle = reverseTitles[from]
	//		er.ToTitle = reverseTitles[to]
	//		er.Count++
	//		edgeRows[ft] = er
	//	}
	//}

	//edgeRowsList := make([]RowEdge, len(edgeRows))
	//for _, er := range edgeRows {
	//	edgeRowsList[idx] = er
	//	idx++
	//}
	//
	//sort.Slice(edgeRowsList, func(i, j int) bool {
	//	return edgeRowsList[i].Count > edgeRowsList[j].Count
	//})
	//
	//lines := make([]string, len(edgeRowsList)+1)
	//lines[0] = "From,To,FromTitle,ToTitle,Count"
	//
	//for i, row := range edgeRowsList {
	//	lines[i+1] = row.String()
	//}

	rowsChan := make(chan Row, jobs.WorkBuffer)

	work := make(chan uint32, jobs.WorkBuffer)

	go func() {
		for id := range ids {
			work <- id
		}

		close(work)
	}()

	wg := sync.WaitGroup{}
	wg.Add(parallel)

	shortestCache := &graphs.ShortestCache{
		Distance: make(map[graphs.FromTo]int, len(graph.Nodes)*len(graph.Nodes)/2),
	}

	for i := 0; i < parallel; i++ {
		go func() {
			for id := range work {
				row := Row{
					ID:        id,
					Title:     reverseTitles[id],
					InDegree:  centrality.InDegree(id, graph),
					OutDegree: centrality.OutDegree(id, graph),
				}

				row.Closeness, row.Harmonic = centrality.ClosenessHarmonic(id, graph, shortestCache)

				rowsChan <- row

				//if id == 41272244 {
				//	cycle := graphs.FindPath(id, id, *graph)
				//	if len(cycle) > 0 {
				//		fmt.Printf("Found cycle for %q\n", reverseTitles[id])
				//
				//		for _, j := range cycle {
				//			fmt.Println(reverseTitles[j])
				//		}
				//
				//		fmt.Println()
				//	}
				//}
			}

			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(rowsChan)
	}()

	rows := make([]Row, len(ids))
	idx := 0

	for row := range rowsChan {
		fmt.Println(idx)
		rows[idx] = row
		idx++
	}


	//for id := range ids {
	//	cycle := graphs.FindPath(id, id, *graph)
	//	if len(cycle) == 0 {
	//		continue
	//	}
	//
	//	for _, n := range cycle {
	//		rows[rowID[n]].Cycle++
	//	}
	//}

	fmt.Println("Sorting")

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Closeness != rows[j].Closeness {
			return rows[i].Closeness > rows[j].Closeness
		}
		return rows[i].Title < rows[j].Title
	})

	lines := make([]string, len(rows)+1)
	lines[0] = "ID,Title,InDegree,OutDegree,Closeness,Harmonic,Cycle"

	for i, row := range rows {
		lines[i+1] = row.String()
	}

	err = os.MkdirAll(filepath.Dir(outCentrality), os.ModePerm)
	if err != nil {
		return err
	}

	outBytes := []byte(strings.Join(lines, "\n"))

	err = ioutil.WriteFile(outCentrality, outBytes, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

type Row struct {
	ID        uint32
	Title     string
	InDegree  int
	OutDegree int
	Closeness float64
	Harmonic  float64
	Cycle int
}

func (r *Row) String() string {
	return fmt.Sprintf("%d,%q,%d,%d,%.3f,%.3f,%d", r.ID, r.Title, r.InDegree, r.OutDegree, r.Closeness, r.Harmonic, r.Cycle)
}

type FromTo struct {
	From uint32
	To   uint32
}

type RowEdge struct {
	FromTo
	FromTitle string
	ToTitle   string
	Count     int
}

func (r *RowEdge) String() string {
	return fmt.Sprintf("%d,%d,%s,%s,%d", r.From, r.To, r.FromTitle, r.ToTitle, r.Count)
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
