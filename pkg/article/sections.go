package article

import (
	"regexp"
	"strings"
)

var HeaderPattern = regexp.MustCompile("\n==[^\n]+==")

type Header struct {
	Text  string
	Level int
}

func (t Header) Render() string {
	return t.Text
}

func ParseHeader(s string) Token {
	nEquals := 2
	for s[nEquals+1] == '=' && s[len(s)-1-nEquals] == '=' && nEquals < 6 {
		nEquals++
	}

	text := s[nEquals+1 : len(s)-nEquals]
	text = strings.TrimSpace(text)
	return Header{
		Text:  text,
		Level: nEquals,
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
	if ignoredSection[s.Header.Text] {
		return ""
	}

	sb := strings.Builder{}

	sb.WriteString("\n")
	sb.WriteString(s.Header.Render())
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
