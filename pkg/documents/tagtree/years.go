package tagtree

import "regexp"

var (
	titleYear      = regexp.MustCompile(`\b\d{3,4}\b`)
	titleYearRange = regexp.MustCompile(`\d{4}–\d{2,4}`)
)

// NodeTitleYear represents a category which programmatically refers to the year in the article's title.
type NodeTitleYear struct{}

func (n *NodeTitleYear) String(title string) string {
	switch matches := titleYear.FindAllString(title, -1); len(matches) {
	case 0:
		return "<YEAR NOT FOUND>"
	case 1:
		return matches[0]
	default:
		return "<MULTIPLE YEARS FOUND>"
	}
}

// NodeTitleYearRange represents a category which programmatically refers to a range of years in the article's title.
type NodeTitleYearRange struct{}

func (n *NodeTitleYearRange) String(title string) string {
	switch matches := titleYearRange.FindAllString(title, -1); len(matches) {
	case 0:
		return "<YEAR RANGE NOT FOUND>"
	case 1:
		return matches[0]
	default:
		return "<MULTIPLE YEAR RANGES FOUND>"
	}
}
