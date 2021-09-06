package tagtree

import (
	"regexp"
	"strings"
)

type NodeMonth struct {
	Value Node
}

func (n *NodeMonth) String(title string) string {
	return n.Value.String(title)
}

var patternMonths = regexp.MustCompile(`(?i)(january|february|march|april|may|june|july|august|september|october|november|december)`)

type NodeTitleMonth struct{}

func (n *NodeTitleMonth) String(title string) string {
	switch months := patternMonths.FindAllString(title, -1); len(months) {
	case 0:
		return "<MONTH NOT FOUND>"
	case 1:
		return months[0]
	default:
		return "<MULTIPLE MONTHS FOUND>"
	}
}

type NodeMonthNumber struct {
	Value Node
}

func (n *NodeMonthNumber) String(title string) string {
	switch strings.ToLower(n.Value.String(title)) {
	case "january":
		return "1"
	case "february":
		return "2"
	case "march":
		return "3"
	case "april":
		return "4"
	case "may":
		return "5"
	case "june":
		return "6"
	case "july":
		return "7"
	case "august":
		return "8"
	case "september":
		return "9"
	case "october":
		return "10"
	case "november":
		return "11"
	case "december":
		return "12"
	}

	return "<MONTH NUMBER NOT FOUND>"
}
