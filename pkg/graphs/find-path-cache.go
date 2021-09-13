package graphs

func FindDistance(start, end uint32, graph Directed, cache *ShortestCache) int {
	q := NewQueue(start)
	visited := map[uint32]uint32{}

	for child := range graph.Nodes[start] {
		cache.Add(start, child, 1)
	}

	var next uint32
	for !q.Empty() {
		next, q = q.Dequeue()

		children, ok := graph.Nodes[next]
		if !ok {
			continue
		}

		for child := range children {
			if next == child {
				// Ignore self-loops.
				continue
			}

			seenDist, seen := cache.Get(child, end)

			if child == end || seen {
				visited[child] = next

				dist := len(unroll(start, child, visited)) + seenDist

				cache.Add(start, end, dist)

				return dist
			}

			if _, isVisited := visited[child]; isVisited {
				continue
			}

			visited[child] = next

			cache.AddNext(start, next, child)

			q = q.Enqueue(child)
		}
	}

	return 0
}
