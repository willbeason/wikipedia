package nlp

import (
	"fmt"
	"regexp"
	"strings"
)

// ignoredTags are tags we can safely strip out, retaining the contents.
func ignoredTags() []string {
	return []string{
		"abbr",
		"big",
		"blockquote",
		"center",
		"div",
		"dfn",
		"em",
		"i",
		"kbd",
		"li",
		"(no|only)?include(only)?",
		"ol",
		"mapframe",
		"nowiki",
		"ref",
		"s",
		"small",
		"span",
		"su[bp]",
		"tt",
		"u",
		"var",
	}
}

// Regular expressions for cleaning Wikipedia articles of XML tags and formatting.
var (

	// XMLTagRegex tries to find XML tags which are still present in the corpus. Useful for finding
	// problematic tags that we want to avoid.
	XMLTagRegex = regexp.MustCompile(`<[a-z]{2,}`)

	// CommentRegex matches commented-out text. Such text is not shown on pages
	// and is generally either off-topic or low quality.
	//
	// Obviously not perfect and can match non-comments in rare cases.
	CommentRegex = regexp.MustCompile("(?s)<!--.*?-->")

	IgnoredTagsRegex     = regexp.MustCompile(fmt.Sprintf(`</?(%s).*?>`, strings.Join(ignoredTags(), "|")))
	TimelineRegex        = regexp.MustCompile(`(?s)<timeline.*?</timeline>`)
	GalleryRegex         = regexp.MustCompile(`(?s)<gallery.*?</gallery>`)
	ImageMapRegex        = regexp.MustCompile(`(?s)<imagemap.*?</imagemap>`)
	MathRegex            = regexp.MustCompile(`(?s)<math.*?</math>`)
	CodeRegex            = regexp.MustCompile(`(?s)<code.*?</code>`)
	ChemRegex            = regexp.MustCompile(`(?s)<chem.*?</chem>`)
	PoemRegex            = regexp.MustCompile(`(?s)<poem.*?</poem>`)
	HieroglyphRegex      = regexp.MustCompile(`(?s)<hiero.*?</hiero>`)
	SyntaxHighlightRegex = regexp.MustCompile(`(?s)<syntaxhighlight.*?</syntaxhighlight>`)
	PreRegex             = regexp.MustCompile(`(?s)<pre.*?</pre>`)
	BrRegex              = regexp.MustCompile(`<(p|br).*?>`)

	AlteredQuote = regexp.MustCompile(`\[([a-zA-Z])]`)

	link = regexp.MustCompile(`\[http[^]]+]`)

	RemoveLinks = regexp.MustCompile(`\[\[(:?Category:|List of)[^]]+]]`)

	WikipediaLinks = regexp.MustCompile(`\[\[([^[\]]+\|)?([^[|]+?)]]`)

	widgets = regexp.MustCompile(`{[^{}]*}`)

	RefRegex = regexp.MustCompile(`(?s)<ref.*?(>.*?</ref>| ?/>)`)
)

// keepReplacing replaces pattern in text with replace until the length of the string stops changing.
func keepReplacing(pattern *regexp.Regexp, text, replace string) string {
	prevLen := len(text)
	text = pattern.ReplaceAllString(text, replace)
	nextLen := len(text)

	for prevLen != nextLen {
		prevLen = nextLen

		text = pattern.ReplaceAllString(text, replace)
		nextLen = len(text)
	}

	return text
}

// CleanArticle removes all parts of Wikipedia we never want to analyze.
func CleanArticle(text string) string {
	text = keepReplacing(widgets, text, "")
	text = RemoveLinks.ReplaceAllString(text, "")

	text = link.ReplaceAllString(text, "")
	text = keepReplacing(WikipediaLinks, text, "$2")

	text = RefRegex.ReplaceAllString(text, "")

	text = CommentRegex.ReplaceAllString(text, "")
	text = GalleryRegex.ReplaceAllString(text, "")
	text = TimelineRegex.ReplaceAllString(text, "")
	text = MathRegex.ReplaceAllString(text, MathToken)
	text = HieroglyphRegex.ReplaceAllString(text, HieroglyphToken)
	text = CodeRegex.ReplaceAllString(text, "")
	text = ChemRegex.ReplaceAllString(text, "")
	text = ImageMapRegex.ReplaceAllString(text, "")
	text = SyntaxHighlightRegex.ReplaceAllString(text, "")
	text = PreRegex.ReplaceAllString(text, "")
	text = PoemRegex.ReplaceAllString(text, "")

	text = IgnoredTagsRegex.ReplaceAllString(text, "")

	text = BrRegex.ReplaceAllString(text, "\n")
	text = AlteredQuote.ReplaceAllString(text, "$1")

	lines := strings.Split(text, "\n")
	result := make([]string, 0, len(lines))
	lastLineEmpty := false

	skip := false

	for _, line := range lines {
		line = strings.ReplaceAll(line, "&nbsp;", " ")
		line = strings.ReplaceAll(line, "&ndash;", "–")

		line = strings.TrimPrefix(line, "*")
		line = strings.TrimSpace(line)

		if line == "==Bibliography==" {
			skip = true
		} else if strings.HasPrefix(line, "==") {
			skip = false
		}

		if skip {
			continue
		}

		if line == "" {
			if !lastLineEmpty {
				result = append(result, line)
			}

			lastLineEmpty = true

			continue
		}

		lastLineEmpty = false

		result = append(result, line)
	}

	text = strings.Join(result, "\n")
	text = strings.TrimSpace(text)

	return text
}
