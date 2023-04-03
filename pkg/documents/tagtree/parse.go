package tagtree

import (
	"errors"
	"regexp"
	"strings"
)

var (
	openTag  = regexp.MustCompile(`{{`)
	closeTag = regexp.MustCompile(`}}`)
	dash     = "dash"
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

var ErrOpenClose = errors.New("different number of open and close braces")

func toBraces(s string) ([]Brace, error) {
	openTags := openTag.FindAllStringIndex(s, -1)
	closeTags := closeTag.FindAllStringIndex(s, -1)

	nTags := len(openTags)
	if nTags == 0 {
		return nil, nil
	}

	if nTags != len(closeTags) {
		return nil, ErrOpenClose
	}

	result := make([]Brace, 0, nTags*2)

	openIDx := 0
	closeIDx := 0

	for openIDx != nTags && closeIDx != nTags {
		if openTags[openIDx][0] < closeTags[closeIDx][0] {
			result = append(result, toBrace(BraceOpen, openTags[openIDx]))
			openIDx++
		} else {
			result = append(result, toBrace(BraceClose, closeTags[closeIDx]))
			closeIDx++
		}
	}

	for openIDx != nTags {
		result = append(result, toBrace(BraceOpen, openTags[openIDx]))
		openIDx++
	}

	for closeIDx != nTags {
		result = append(result, toBrace(BraceClose, closeTags[closeIDx]))
		closeIDx++
	}

	return result, nil
}

func Parse(category string) Node {
	braces, err := toBraces(category)
	if err != nil {
		return &NodeError{Value: err}
	}

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

	sIDx := braces[0].Start
	parent := &NodeParent{
		Children: []Node{
			&NodeString{Value: category[:sIDx]},
		},
	}

	level := 0

	for i, brace := range braces {
		switch brace.Type {
		case BraceOpen:
			if level == 0 {
				if sIDx != braces[i].Start {
					parent.Children = append(parent.Children, &NodeString{Value: category[sIDx:braces[i].Start]})
				}

				sIDx = braces[i].Start
			}

			level++
		case BraceClose:
			level--

			if level == 0 {
				parent.Children = append(parent.Children, parseTag(category[sIDx:braces[i].End]))
				sIDx = braces[i].End
			} else if level < 0 {
				return &NodeString{Value: category}
			}
		default:
			return &NodeString{Value: category}
		}
	}

	if sIDx != len(category) {
		parent.Children = append(parent.Children, &NodeString{Value: category[sIDx:]})
	}

	return parent
}

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

		dash := splits[len(splits)-1] == dash

		return &NodeCentury{
			Value: Parse(splits[1]),
			Dash:  dash,
		}
	case "century name from title year":
		dash := splits[len(splits)-1] == dash

		return &NodeCentury{Value: &NodeTitleYear{}, Dash: dash}
	case "century name from title decade":
		dash := splits[len(splits)-1] == dash

		return &NodeCentury{Value: &NodeTitleDecade{}, Dash: dash}
	case "century name from decade or year":
		dash := splits[len(splits)-1] == dash

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
	case "first word":
		if len(splits) == 1 {
			return &NodeFirstWord{Value: &NodeString{"<MISSING FIRST WORD NODE>"}}
		}

		return &NodeFirstWord{Value: Parse(splits[1])}
	case "last word":
		if len(splits) == 1 {
			return &NodeLastWord{Value: &NodeString{"<MISSING LAST WORD NODE>"}}
		}

		return &NodeLastWord{Value: Parse(splits[1])}
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
		return &NodeString{Value: category}
	}
}
