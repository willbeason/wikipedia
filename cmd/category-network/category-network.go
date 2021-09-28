package main

import (
	"context"
	"fmt"
	"io/ioutil"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	"github.com/willbeason/wikipedia/pkg/classify"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/graphs"
	"github.com/willbeason/wikipedia/pkg/jobs"
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
	inKnown := args[2]
	outClassifications := args[3]

	knownClassifications := &classify.ClassifiedTitles{}
	err = protos.Read(inKnown, knownClassifications)
	if err != nil {
		return err
	}

	pageCategories := &documents.PageCategories{}
	err = protos.Read(inCategories, pageCategories)
	if err != nil {
		return err
	}

	pageTitles := &documents.TitleIndex{}
	err = protos.Read(inTitles, pageTitles)
	if err != nil {
		return err
	}

	reverseTitles := make(map[uint32]string)
	for title, id := range pageTitles.Titles {
		reverseTitles[id] = title
	}

	ignoredArticlesBytes, err := ioutil.ReadFile("./data/ignored-articles.txt")
	if err != nil {
		return err
	}
	ignoredArticles := strings.Split(string(ignoredArticlesBytes), "\n")
	ignoredArticlesMap := make(map[uint32]bool, len(ignoredArticles))
	for _, title := range ignoredArticles {
		title = strings.TrimSpace(title)
		if title == "" {
			continue
		}

		id, ok := pageTitles.Titles[title]
		if !ok {
			panic(fmt.Sprintf("%q", title))
		}

		ignoredArticlesMap[id] = true
	}

	// known2 := knownClassifications.ToClassifiedIDs(pageTitles.Titles).Pages

	graph := make(map[uint32]map[uint32]bool, len(pageCategories.Pages))
	for page, categories := range pageCategories.Pages {
		if ignoredArticlesMap[page] {
			continue
		}

		edges := make(map[uint32]bool, len(categories.Categories))
		for _, c := range categories.Categories {
			if ignoredArticlesMap[c] {
				continue
			}

			edges[c] = true
		}

		graph[page] = edges
	}

	connectedGraphs := findAllConnectedDigraphs(parallel, graphs.Directed{Nodes: graph})

	err = os.MkdirAll(outClassifications, os.ModePerm)
	if err != nil {
		return err
	}
	for i, g := range connectedGraphs {
		titles := make([]string, len(g))

		j := 0
		for n := range g {
			titles[j] = reverseTitles[n]
			j++
		}

		sort.Strings(titles)
		bytes := []byte(strings.Join(titles, "\n"))

		file := filepath.Join(outClassifications, fmt.Sprintf("%03d.txt", i))
		err = ioutil.WriteFile(file, bytes, os.ModePerm)
		if err != nil {
			return err
		}
		fmt.Println("Wrote to", file)
	}

	fmt.Println(len(connectedGraphs))

	// classifications := buildMap(known2, reverseTitles, pageCategories)

	// Post-classification analysis.

	//all := 0
	//alls := make(chan uint32, 100)
	//
	//go func() {
	//	solos := make(map[classify.Classification]uint32)
	//
	//	for id, c := range classifications.Pages {
	//		cl := onlyUnknown(c.Classifications)
	//		if cl == nil {
	//
	//			nHas := 0
	//			for i, cc := range c.Classifications {
	//				if i == int(cc) {
	//					nHas++
	//					if nHas >= 14 {
	//						break
	//					}
	//				}
	//			}
	//			if nHas >= 14 {
	//				alls <- id
	//				all++
	//			}
	//
	//			continue
	//		}
	//		solos[*cl]++
	//	}
	//	close(alls)
	//
	//	for i := int32(0); i < 20; i++ {
	//		fmt.Println(classify.Classification_name[i], solos[classify.Classification(i)])
	//	}
	//	fmt.Println("ALL", all)
	//	fmt.Println()
	//}()
	//
	//shortestWg := sync.WaitGroup{}
	//shortestWg.Add(parallel)
	//
	//paths := make(chan []uint32)
	//
	//for i := 0; i < parallel; i++ {
	//	go func() {
	//		for id := range alls {
	//			for c := classify.Classification_PHILOSOPHY; c <= classify.Classification_INFORMATION_SCIENCE; c++ {
	//				paths <- shortestTo(id, c, pageCategories, known2)
	//			}
	//		}
	//
	//		shortestWg.Done()
	//	}()
	//}
	//
	//go func() {
	//	shortestWg.Wait()
	//	close(paths)
	//}()
	//
	//nodeFrequency := make(map[uint32]uint32)
	//
	//fmt.Println("Combining paths")
	//nCombined := 0
	//
	//for path := range paths {
	//	for _, id := range path {
	//		nodeFrequency[id]++
	//	}
	//	nCombined++
	//	if nCombined%1000 == 0 {
	//		fmt.Println("Combined", nCombined)
	//	}
	//}
	//
	//fmt.Println("Combined paths")
	//
	//sortedList := make([]f, len(nodeFrequency))
	//idx := 0
	//
	//for id, count := range nodeFrequency {
	//	sortedList[idx] = f{id: id, count: count}
	//	idx++
	//}
	//
	//sort.Slice(sortedList, func(i, j int) bool {
	//	return sortedList[i].count > sortedList[j].count
	//})
	//
	//for _, entry := range sortedList[:100] {
	//	if _, isKnown := known2[entry.id]; isKnown {
	//		fmt.Println("KNOWN", reverseTitles[entry.id], ":", entry.count)
	//	} else {
	//		fmt.Println(reverseTitles[entry.id], ":", entry.count)
	//	}
	//}

	// return protos.Write(outClassifications, classifications)
	return nil
}

type f struct {
	id    uint32
	count uint32
}

func buildMap(known map[uint32]classify.Classification, reverseTitles map[uint32]string, pageCategories *documents.PageCategories) *classify.PageClassificationsMap {
	result := &classify.PageClassificationsMap{
		Pages: make(map[uint32]*classify.PageClassifications),
	}

	for id, c := range known {
		result.Pages[id] = coreClassification(c)
	}

	n := 0

	for pageId, categories := range pageCategories.Pages {
		//if pageId != 48628332 {
		//	continue
		//}

		result.AddPage(known, reverseTitles, pageId, categories.Categories, pageCategories)

		n++
		if n%10000 == 0 {
			fmt.Println(n)
		}
	}

	return result
}

func onlyUnknown(cs []classify.Classification) *classify.Classification {
	only := classify.Classification_UNKNOWN
	for _, c := range cs {
		if c != classify.Classification_UNKNOWN {
			if only != classify.Classification_UNKNOWN {
				return nil
			}

			only = c
		}
	}

	return &only
}

func coreClassification(c classify.Classification) *classify.PageClassifications {
	result := make([]classify.Classification, 20)

	result[c] = c

	return &classify.PageClassifications{Classifications: result}
}

func shortestTo(pageId uint32, want classify.Classification, pageCategories *documents.PageCategories, known map[uint32]classify.Classification) []uint32 {
	visited := make(map[uint32]uint32)

	toVisit := map[uint32]bool{
		pageId: true,
	}
	var nextToVisit map[uint32]bool

	depth := 0

	for len(toVisit) > 0 {
		depth++
		nextToVisit = make(map[uint32]bool)

		for k := range toVisit {
			if got, isKnown := known[k]; isKnown {
				if got == want {
					return unwind(k, visited, depth)
				}

				continue
			}

			for _, c := range pageCategories.Pages[k].Categories {
				if _, seen := visited[c]; seen || c == pageId {
					continue
				}

				visited[c] = k

				nextToVisit[c] = true
			}
		}

		toVisit = nextToVisit
	}

	// fmt.Println("NO PATH")
	return nil
}

func unwind(k uint32, visited map[uint32]uint32, depth int) []uint32 {
	result := make([]uint32, depth)
	result[0] = k

	next, found := visited[k]

	for found {
		result = append(result, next)
		next, found = visited[next]
	}

	//for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
	//	result[i], result[j] = result[j], result[i]
	//}

	//fmt.Println("Length:", len(result))
	//for _, id := range result {
	//	fmt.Println(reverseTitles[id])
	//}

	return result
}

func findAllConnectedDigraphs(parallel int, graph graphs.Directed) []map[uint32]bool {
	inCycles := make(map[uint32]bool)
	inCyclesMtx := sync.RWMutex{}

	connectedGraphsChan := make(chan []uint32, jobs.WorkBuffer)

	nProcessed := 0
	nMtx := sync.Mutex{}

	work := make(chan uint32, jobs.WorkBuffer)

	go func() {
		for k := range graph.Nodes {
			work <- k

			nMtx.Lock()
			nProcessed++
			if nProcessed%10000 == 0 {
				fmt.Println(nProcessed)
			}
			nMtx.Unlock()
		}
		close(work)
	}()

	wg := sync.WaitGroup{}
	wg.Add(parallel)

	for i := 0; i < parallel; i++ {
		go func() {
			for k := range work {
				inCyclesMtx.RLock()
				seen := inCycles[k]
				inCyclesMtx.RUnlock()
				if seen {
					continue
				}

				cycle := graphs.FindCycle(k, graph)

				if len(cycle) == 0 {
					continue
				}

				inCyclesMtx.Lock()
				for _, node := range cycle {
					inCycles[node] = true
				}
				inCyclesMtx.Unlock()

				connectedGraphsChan <- cycle
			}

			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(connectedGraphsChan)
	}()

	var connectedGraphs []map[uint32]bool

	for g := range connectedGraphsChan {
		cg := make(map[uint32]bool, len(g))
		for _, k := range g {
			cg[k] = true
		}
		connectedGraphs = append(connectedGraphs, cg)
	}

	fmt.Println("Found", len(connectedGraphs), "cycles")

	for i := 0; i < len(connectedGraphs); i++ {
		iStart := getKey(connectedGraphs[i])
		if iStart == nil {
			continue
		}

		for j := i + 1; j < len(connectedGraphs); {
			jStart := getKey(connectedGraphs[j])
			if jStart == nil {
				continue
			}

			to := graphs.FindPath(*iStart, *jStart, graph)
			if len(to) == 0 {
				j++
				continue
			}

			from := graphs.FindPath(*jStart, *iStart, graph)
			if len(from) == 0 {
				j++
				continue
			}

			connectedGraphs = merge(i, j, connectedGraphs)
			fmt.Println("Merged", i, "and", j, ". ", len(connectedGraphs), "connected graphs remaining.")
		}
	}

	sort.Slice(connectedGraphs, func(i, j int) bool {
		return len(connectedGraphs[i]) > len(connectedGraphs[j])
	})

	fmt.Println("Found", len(connectedGraphs), "connected graphs")
	if len(connectedGraphs) > 0 {
		fmt.Println("Largest connected graph is of size", len(connectedGraphs[0]))
	}

	return connectedGraphs
}

func getKey(m map[uint32]bool) *uint32 {
	for k := range m {
		return &k
	}
	return nil
}

func merge(x, y int, gs []map[uint32]bool) []map[uint32]bool {
	gsx := gs[x]
	for k := range gs[y] {
		gsx[k] = true
	}
	gs[x] = gsx

	copy(gs[y:], gs[y+1:])
	return gs[:len(gs)-1]
}
