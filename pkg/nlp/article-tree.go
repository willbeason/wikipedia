package nlp

import (
	"strings"
)

const (
	// MinSectionDepth is the lowest section depth in wikitext by convention.
	// A depth of 1 is technically allowed but causes formatting errors.
	MinSectionDepth = 2
	// MaxSectionDepth is the highest section depth wikitext recognizes.
	MaxSectionDepth = 6
)

func isHeader(line string) bool {
	return strings.HasPrefix(line, "==") &&
		strings.HasSuffix(line, "==")
}

// ParseText parses lines of a Wikipedia article into a tree representing
// the article's wikitext structure.
func ParseText(lines []string) TextNode {
	// Check if this is a section.
	if isHeader(lines[0]) {
		return ParseSection(lines)
	}

	result := &TextTree{}

	// Text is not a section, but might contain sections.
	sectionStart := 0
	seenSection := false
	for sectionEnd := 1; sectionEnd <= len(lines); sectionEnd++ {
		// Treat as a section if we see a header, or we reach the end of the
		// text to parse and have previously seen a section.
		if isHeader(lines[sectionEnd]) || (sectionEnd == len(lines) && seenSection) {
			result.Children = append(result.Children,
				ParseText(lines[sectionStart:sectionEnd]))
			sectionStart = sectionEnd
			seenSection = true
		}
	}

	// We parsed this text into sections; don't mix except the first set of text.
	if seenSection {
		return result
	}

	// Text is just lines and contains no section headers.
	for _, line := range lines {
		result.Children = append(result.Children, TextLiteral(line))
	}

	return result
}

func ParseSection(lines []string) *SectionLiteral {
	result := &SectionLiteral{}

	result.Heading = strings.Trim(lines[0], "= ")
	result.Text = ParseText(lines[1:])

	return result
}

type TextTree struct {
	Children []TextNode
}

func (t TextTree) Print() string {
	sb := strings.Builder{}

	for _, child := range t.Children {
		sb.WriteString(child.Print())
	}

	return sb.String()
}

type TextNode interface {
	Print() string
}

type TextLiteral string

func (t TextLiteral) Print() string {
	return string(t)
}

type SectionLiteral struct {
	Heading string
	Text    TextNode
}

func (s SectionLiteral) Print() string {
	return s.Text.Print()
}
