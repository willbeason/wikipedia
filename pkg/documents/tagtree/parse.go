package tagtree

import (
	"regexp"
	"strings"
)

var (
	openTag  = regexp.MustCompile(`{{`)
	closeTag = regexp.MustCompile(`}}`)
)

type Brace struct {
	Type       BraceType
	Start, End int
}

type BraceType int

const (
	BraceOpen BraceType = iota
	BraceClose
)

func toBrace(t BraceType, startEnd []int) Brace {
	return Brace{Type: t, Start: startEnd[0], End: startEnd[1]}
}

func toBraces(s string) []Brace {
	openTags := openTag.FindAllStringIndex(s, -1)
	closeTags := closeTag.FindAllStringIndex(s, -1)

	nTags := len(openTags)
	if nTags == 0 || nTags != len(closeTags) {
		return nil
	}

	result := make([]Brace, 0, nTags*2)

	openIdx := 0
	closeIdx := 0

	for openIdx != nTags && closeIdx != nTags {
		if openTags[openIdx][0] < closeTags[closeIdx][0] {
			result = append(result, toBrace(BraceOpen, openTags[openIdx]))
			openIdx++
		} else {
			result = append(result, toBrace(BraceClose, closeTags[closeIdx]))
			closeIdx++
		}
	}

	for openIdx != nTags {
		result = append(result, toBrace(BraceOpen, openTags[openIdx]))
		openIdx++
	}

	for closeIdx != nTags {
		result = append(result, toBrace(BraceClose, closeTags[closeIdx]))
		closeIdx++
	}

	return result
}

func Parse(category string) Node {
	braces := toBraces(category)

	if len(braces) == 0 {
		return &NodeString{Value: category}
	}

	if braces[0].Type != BraceOpen || braces[len(braces)-1].Type != BraceClose {
		return &NodeString{Value: category}
	}

	if len(braces) == 2 {
		result := &NodeParent{}

		if braces[0].Start != 0 {
			result.Children = append(result.Children, &NodeString{Value: category[:braces[0].Start]})
		}

		result.Children = append(result.Children, parseTag(category[braces[0].Start:braces[1].End]))

		if braces[1].End < len(category) {
			result.Children = append(result.Children, &NodeString{Value: category[braces[1].End:]})
		}

		if len(result.Children) == 1 {
			return result.Children[0]
		}

		return result
	}

	sIdx := braces[0].Start
	parent := &NodeParent{
		Children: []Node{
			&NodeString{Value: category[:sIdx]},
		},
	}

	level := 0

	for i, brace := range braces {
		switch brace.Type {
		case BraceOpen:
			if level == 0 {
				if sIdx != braces[i].Start {
					parent.Children = append(parent.Children, &NodeString{Value: category[sIdx:braces[i].Start]})
				}

				sIdx = braces[i].Start
			}

			level++
		case BraceClose:
			level--

			if level == 0 {
				parent.Children = append(parent.Children, parseTag(category[sIdx:braces[i].End]))
				sIdx = braces[i].End
			} else if level < 0 {
				return &NodeString{Value: category}
			}
		default:
			return &NodeString{Value: category}
		}
	}

	if sIdx != len(category) {
		parent.Children = append(parent.Children, &NodeString{Value: category[sIdx:]})
	}

	return parent
}

var MissedMap = map[string]int{}

func parseTag(category string) Node {
	category = strings.TrimPrefix(category, "{{")
	category = strings.TrimSuffix(category, "}}")
	category = strings.TrimSpace(category)

	if strings.HasPrefix(category, "#expr:") {
		return &NodeExpression{Value: Parse(category[6:])}
	}

	splits := strings.Split(category, "|")

	tagType := splits[0]
	tagType = strings.TrimSpace(tagType)

	switch strings.ToLower(tagType) {
	case "title year":
		return &NodeTitleYear{}
	case "title year range":
		return &NodeTitleYearRange{}
	case "title monthname":
		return &NodeTitleMonth{}
	case "title country":
		return &NodeTitleCountry{}
	case "title decade":
		return &NodeTitleDecade{}
	case "title century":
		return &NodeTitleCentury{}
	case "country2continent":
		if len(splits) == 1 {
			return &NodeCountry2Continent{Value: &NodeString{Value: "<MISSING COUNTRY2CONTINENT NODE>"}}
		}

		return &NodeCountry2Continent{Value: Parse(splits[1])}
	case "country2nationality":
		if len(splits) == 1 {
			return &NodeCountry2Nationality{Value: &NodeString{Value: "<MISSING COUNTRY2CONTINENT NODE>"}}
		}

		return &NodeCountry2Nationality{Value: Parse(splits[1])}
	case "continent2continental":
		if len(splits) == 1 {
			return &NodeContinent2Continental{Value: &NodeString{Value: "<MISSING CONTINENT NODE>"}}
		}

		return &NodeContinent2Continental{Value: Parse(splits[1])}
	case "country2continental":
		if len(splits) == 1 {
			return &NodeContinent2Continental{Value: &NodeString{Value: "<MISSING CONTINENT NODE>"}}
		}

		return &NodeContinent2Continental{Value: &NodeCountry2Continent{Value: Parse(splits[1])}}
	case "century from year":
		if len(splits) == 1 {
			return &NodeCentury{Value: &NodeString{Value: "<MISSING CENTURY NODE>"}}
		}

		dash := splits[len(splits)-1] == "dash"
		return &NodeCentury{
			Value: Parse(splits[1]),
			Dash:  dash,
		}
	case "century name from title year":
		dash := splits[len(splits)-1] == "dash"
		return &NodeCentury{Value: &NodeTitleYear{}, Dash: dash}
	case "century name from title decade":
		dash := splits[len(splits)-1] == "dash"
		return &NodeCentury{Value: &NodeTitleDecade{}, Dash: dash}
	case "century name from decade or year":
		dash := splits[len(splits)-1] == "dash"

		if len(splits) == 1 {
			return &NodeCentury{Value: &NodeString{Value: "<MISSING CENTURY NODE>"}}
		}

		return &NodeCentury{Value: Parse(splits[1]), Dash: dash}
	case "decade":
		if len(splits) == 1 {
			return &NodeDecade{Value: &NodeString{Value: "<MISSING DECADE NODE>"}}
		}

		return &NodeDecade{Value: Parse(splits[1])}
	case "month":
		if len(splits) == 1 {
			return &NodeMonth{Value: &NodeString{Value: "<MISSING MONTH NODE>"}}
		}

		return &NodeMonth{Value: Parse(splits[1])}
	case "ordinal":
		if len(splits) == 1 {
			return &NodeOrdinal{Value: &NodeString{Value: "<MISSING ORDINAL NODE>"}}
		}

		return &NodeOrdinal{Value: Parse(splits[1])}
	case "pagename":
		return &NodePageName{}
	case "currentyear":
		return &NodeString{Value: "2021"}
	case "monthnumber":
		if len(splits) == 1 {
			return &NodeOrdinal{Value: &NodeString{"<MISSING MONTH NUMBER NODE>"}}
		}

		return &NodeMonthNumber{Value: Parse(splits[1])}
	case "title year+1":
		return &NodeExpression{Value: &NodeParent{Children: []Node{
			&NodeTitleYear{},
			&NodeString{Value: "+1"},
		}}}
	case "title year-1":
		return &NodeExpression{Value: &NodeParent{Children: []Node{
			&NodeTitleYear{},
			&NodeString{Value: "-1"},
		}}}
	default:
		MissedMap[splits[0]]++
		return &NodeString{Value: category}
	}
}
