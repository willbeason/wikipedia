package tagtree

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type NodeCentury struct {
	Value Node
	Dash  bool
}

func (n *NodeCentury) String(title string) string {
	year := n.Value.String(title)

	var century uint64

	if _, isDecade := n.Value.(*NodeTitleDecade); isDecade || strings.HasSuffix(year, "s") {
		year = strings.TrimSuffix(year, "s")

		nYear, err := strconv.ParseUint(year, 10, 16)
		if err != nil {
			return "<UNABLE TO PARSE DECADE>"
		}

		century = nYear / 100
		century++
	} else {
		nYear, err := strconv.ParseUint(year, 10, 16)
		if err != nil {
			return "<UNABLE TO PARSE YEAR>"
		}

		century = nYear / 100
		if nYear%100 != 0 {
			century++
		}
	}

	ordinalCentury := ordinal(century)

	if n.Dash {
		return ordinalCentury + "-century"
	}

	return ordinalCentury + " century"
}

func ordinal(n uint64) string {
	if (n%100)/10 == 1 {
		return fmt.Sprintf("%dth", n)
	}

	switch n % 10 {
	case 1:
		return fmt.Sprintf("%dst", n)
	case 2:
		return fmt.Sprintf("%dnd", n)
	case 3:
		return fmt.Sprintf("%drd", n)
	default:
		return fmt.Sprintf("%dth", n)
	}
}

var patternCentury = regexp.MustCompile(`(?i)(\d+)(st|nd|rd|th)[- ]century`)

type NodeTitleCentury struct{}

func (n *NodeTitleCentury) String(title string) string {
	switch centuries := patternCentury.FindAllStringSubmatch(title, -1); len(centuries) {
	case 0:
		return "<CENTURY NOT FOUND>"
	case 1:
		return centuries[0][1]
	default:
		return "<MULTIPLE CENTURIES FOUND>"
	}
}

type NodeOrdinal struct {
	Value Node
}

func (n *NodeOrdinal) String(title string) string {
	v := n.Value.String(title)

	vn, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return "<UNABLE TO PARSE CENTURY>"
	}

	return ordinal(vn)
}
