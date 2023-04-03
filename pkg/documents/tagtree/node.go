package tagtree

import (
	"strings"
)

// Node is a part of a referenced Category.
type Node interface {
	String(title string) string
}

// A NodeParent is made of.
type NodeParent struct {
	Children []Node
}

func (n *NodeParent) String(title string) string {
	sb := strings.Builder{}

	for _, c := range n.Children {
		_, _ = sb.WriteString(c.String(title))
	}

	return sb.String()
}

type NodeString struct {
	Value string
}

func (n *NodeString) String(title string) string {
	return n.Value
}

type NodeError struct {
	Value error
}

func (n *NodeError) String(title string) string {
	return n.Value.Error()
}

type NodePageName struct{}

func (n *NodePageName) String(title string) string {
	title = strings.TrimPrefix(title, "Category:")

	return title
}
