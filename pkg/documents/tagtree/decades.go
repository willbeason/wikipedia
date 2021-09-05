package tagtree

import "regexp"

type NodeDecade struct {
	Value Node
}

func (n *NodeDecade) String(title string) string {
	v := n.Value.String(title)
	if v == "" {
		return "<MISSING DECADE>"
	}

	return v[:len(v)-1] + "0s"
}

var patternDecades = regexp.MustCompile(`\d{2,3}0`)

type NodeTitleDecade struct {}

func (n *NodeTitleDecade) String(title string) string {
	switch decades := patternDecades.FindAllString(title, -1); len(decades) {
	case 0:
		return "<DECADE NOT FOUND>"
	case 1:
		return decades[0]
	default:
		return "<MULTIPLE DECADES FOUND>"
	}
}
