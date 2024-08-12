package article

import (
	"regexp"
	"strings"
)

var (
	HeaderStartPattern    = regexp.MustCompile("(?m)^={2,6}")
	HeaderEndPattern      = regexp.MustCompile("(?m)={2,6}$")
	HeaderStartEndPattern = regexp.MustCompile("(?m)^={2,6}$")
)

type HeaderStart string

func (t HeaderStart) Render() string {
	return string(t)
}

func ParseHeaderStart(s string) Token {
	return HeaderStart(s)
}

type HeaderEnd string

func (t HeaderEnd) Render() string {
	return string(t)
}

func ParseHeaderEnd(s string) Token {
	return HeaderEnd(s)
}

func (t HeaderEnd) Backtrack(tokens []Token) (Token, int) {
	start, startIdx, found := BacktrackUntil[HeaderStart](tokens)
	if !found {
		return nil, startIdx
	}

	text := make([]Token, len(tokens)-(startIdx+1))
	copy(text, tokens[startIdx+1:])

	return ParseHeader(start, text, t), startIdx
}

type HeaderStartEnd string

func (t HeaderStartEnd) Render() string {
	return string(t)
}

type Header struct {
	Text  []Token
	Level int
}

func (t Header) Render() string {
	return Render(t.Text)
}

func ParseHeader(start HeaderStart, text []Token, end HeaderEnd) Token {
	startLevel := len(start)
	endLevel := len(end)
	level := startLevel

	if endLevel < startLevel {
		startEquals := strings.Repeat("=", startLevel-endLevel)
		text = append([]Token{LiteralText(startEquals)}, text...)
		// Begin with equals
		level = endLevel
	} else if startLevel < endLevel {
		// End with equals
		endEquals := strings.Repeat("=", endLevel-startLevel)
		text = append(text, LiteralText(endEquals))
		level = startLevel
	}

	return Header{
		Text:  text,
		Level: level,
	}
}

type Section struct {
	Header Header
	Text   []Token
}

// nolint: gochecknoglobals
var ignoredSection = map[string]bool{
	"Articles":           true,
	"External links":     true,
	"Further reading":    true,
	"Notes":              true,
	"Online biographies": true,
	"References":         true,
	"See also":           true,
	"Sources":            true,
}

func (s Section) Render() string {
	renderedHeader := Render(s.Header.Text)
	renderedHeader = strings.TrimSpace(renderedHeader)
	if ignoredSection[renderedHeader] {
		return ""
	}

	sb := strings.Builder{}

	sb.WriteString(renderedHeader)
	for _, text := range s.Text {
		sb.WriteString(text.Render())
	}

	return sb.String()
}

func MergeSections(tokens []Token) []Token {
	var result []Token

	headerIdx := -1
	var header Header
	var sectionTokens []Token

	for idx, token := range tokens {
		startHeader, isHeader := token.(Header)
		if !isHeader {
			if headerIdx == -1 {
				// No previous headers.
				result = append(result, token)
			} else {
				// We are in a section.
				sectionTokens = append(sectionTokens, token)
			}
			continue
		} else if headerIdx != -1 && header.Level < startHeader.Level {
			// We are in a subsection of the current header.
			sectionTokens = append(sectionTokens, token)
			continue
		}

		if headerIdx != -1 {
			// Close off previous header.
			sectionTokens = MergeSections(sectionTokens)

			section := Section{
				Header: header,
				Text:   sectionTokens,
			}
			result = append(result, section)

			sectionTokens = nil
		}

		headerIdx = idx
		header = startHeader
	}

	if headerIdx != -1 {
		// Close off previous header.
		sectionTokens = MergeSections(sectionTokens)

		section := Section{
			Header: header,
			Text:   sectionTokens,
		}
		result = append(result, section)
	}

	return result
}
