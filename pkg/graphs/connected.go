package graphs

type Connected map[uint32]bool

// FindCycle returns a loop containing start in graph, or nil if there is no such loop.
func FindCycle(start uint32, graph Directed) []uint32 {
	if graph.Nodes == nil {
		return nil
	}

	return FindPath(start, start, graph)
}

type Queue struct {
	Value    uint32
	Next     *Queue
	Previous *Queue
}

func NewQueue(value uint32) *Queue {
	q := &Queue{Value: value, Next: nil, Previous: nil}

	q.Next = q
	q.Previous = q

	return q
}

func (q *Queue) Enqueue(value uint32) *Queue {
	if q == nil {
		return NewQueue(value)
	}

	end := &Queue{
		Value:    value,
		Next:     q,
		Previous: q.Previous,
	}

	q.Previous.Next = end
	q.Previous = end

	return q
}

func (q *Queue) Dequeue() (uint32, *Queue) {
	if q == q.Next {
		return q.Value, nil
	}

	q.Previous.Next = q.Next
	q.Next.Previous = q.Previous

	return q.Value, q.Next
}

func (q *Queue) Empty() bool {
	return q == nil
}

func FindPath(start, end uint32, graph Directed) []uint32 {
	toVisit := NewQueue(start)
	visited := map[uint32]uint32{}

	var next uint32
	for !toVisit.Empty() {
		next, toVisit = toVisit.Dequeue()

		children, ok := graph.Nodes[next]
		if !ok {
			continue
		}

		for child := range children {
			if next == child {
				// Ignore self-loops.
				continue
			}

			if child == end {
				visited[child] = next

				return unroll(start, child, visited)
			}

			if _, isVisited := visited[child]; isVisited {
				continue
			}

			visited[child] = next

			toVisit = toVisit.Enqueue(child)
		}
	}

	return nil
}

func unroll(start, end uint32, visited map[uint32]uint32) []uint32 {
	var path []uint32
	if start != end {
		// Since the path is a loop, don't duplicate the end node.
		path = []uint32{end}
	}

	seenInPath := map[uint32]bool{end: true}

	next, found := visited[end]
	for next != start && found {
		if seenInPath[next] {
			break
		}

		seenInPath[next] = true

		path = append(path, next)

		next, found = visited[next]
	}

	path = append(path, start)

	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path
}
