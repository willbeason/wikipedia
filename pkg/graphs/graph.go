package graphs

// Directed is a thread-unsafe representation of a directed Graph.
type Directed struct {
	Nodes map[uint32]map[uint32]bool
}
