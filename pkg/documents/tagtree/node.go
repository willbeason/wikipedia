package tagtree

import (
	"strings"
)

type Node interface {
	String(title string) string
}

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

func(n *NodeString) String(title string) string {
	return n.Value
}

type NodePageName struct {}

func (n *NodePageName) String(title string) string {
	return title
}
