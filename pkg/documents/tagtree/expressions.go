package tagtree

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type NodeExpression struct {
	Value Node
}

var (
	patternPlus  = regexp.MustCompile(`(\d+)\+(\d+)`)
	patternMinus = regexp.MustCompile(`(\d+)-(\d+)`)
)

func (n *NodeExpression) String(title string) string {
	v := n.Value.String(title)

	v = strings.ReplaceAll(v, " ", "")

	switch matches := patternPlus.FindAllStringSubmatch(v, -1); len(matches) {
	case 0:
		// Do nothing
	case 1:
		left, err := strconv.ParseUint(matches[0][1], 10, 64)
		if err != nil {
			return "<UNABLE TO PARSE EXPRESSION LEFT>"
		}

		right, err := strconv.ParseUint(matches[0][2], 10, 64)
		if err != nil {
			return "<UNABLE TO PARSE EXPRESSION RIGHT>"
		}

		return fmt.Sprint(left + right)
	default:
		return "<MULTIPLE PLUS EXPRESSIONS>"
	}

	switch matches := patternMinus.FindAllStringSubmatch(v, -1); len(matches) {
	case 0:
		// Do nothing
	case 1:
		left, err := strconv.ParseUint(matches[0][1], 10, 64)
		if err != nil {
			return "<UNABLE TO PARSE EXPRESSION LEFT>"
		}

		right, err := strconv.ParseUint(matches[0][2], 10, 64)
		if err != nil {
			return "<UNABLE TO PARSE EXPRESSION RIGHT>"
		}

		return fmt.Sprint(left - right)
	default:
		return "<MULTIPLE MINUS EXPRESSIONS>"
	}

	return v
}
