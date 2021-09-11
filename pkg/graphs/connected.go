package graphs

type Connected map[uint32]bool

// FindCycle returns a loop containing start in graph, or nil if there is no such loop.
func FindCycle(start uint32, graph Directed) []uint32 {
	if graph.Nodes == nil {
		return nil
	}

	return FindPath(start, start, graph)
}

func FindPath(start, end uint32, graph Directed) []uint32 {
	return findPath(end, graph, start, []uint32{start}, map[uint32]bool{start: true})
}

func findPath(end uint32, graph Directed, next uint32, stack []uint32, visited map[uint32]bool) []uint32 {
	children, ok := graph.Nodes[next]
	if !ok {
		return nil
	}

	for child := range children {
		if visited[child] {
			// This explicitly ignores self-loops.
			continue
		}

		if child == end {
			// We found a path!
			return stack
		}

		visited[child] = true

		loop := findPath(end, graph, child, append(stack, child), visited)
		if loop != nil {
			return loop
		}
	}

	return nil
}
