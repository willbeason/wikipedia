package tagtree

import "strings"

type NodeFirstWord struct {
	Value Node
}

func (n *NodeFirstWord) String(title string) string {
	v := n.Value.String(title)

	return strings.Split(v, " ")[0]
}

type NodeLastWord struct {
	Value Node
}

func (n *NodeLastWord) String(title string) string {
	v := n.Value.String(title)
	splits := strings.Split(v, " ")

	return splits[len(splits)-1]
}
