package nlp

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	IgnoredTags = []string{
		"abbr",
		"big",
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
		"poem",
		"ref",
		"s",
		"small",
		"span",
		"su[bp]",
		"tt",
		"u",
		"var",
	}

	// XMLTagRegex tries to find XML tags which are still present in the corpus. Useful for finding
	// problematic tags that we want to avoid.
	XMLTagRegex = regexp.MustCompile(`<[a-z]{2,}`)

	// CommentRegex matches commented-out text. Such text is not shown on pages
	// and is generally either off-topic or low quality.
	//
	// Obviously not perfect and can match non-comments in rare cases.
	CommentRegex = regexp.MustCompile("(?s)<!--.*?-->")

	IgnoredTagsRegex     = regexp.MustCompile(fmt.Sprintf(`</?(%s).*?>`, strings.Join(IgnoredTags, "|")))
	SubSupRegex          = regexp.MustCompile(`</?su[bp].*?>`)
	BlockQuoteRegex      = regexp.MustCompile(`</?blockquote.*?>`)
	TimelineRegex        = regexp.MustCompile(`(?s)<timeline.*?</timeline>`)
	GalleryRegex         = regexp.MustCompile(`(?s)<gallery.*?</gallery>`)
	ImageMapRegex        = regexp.MustCompile(`(?s)<imagemap.*?</imagemap>`)
	MathRegex            = regexp.MustCompile(`(?s)<math.*?</math>`)
	CodeRegex            = regexp.MustCompile(`(?s)<code.*?</code>`)
	ChemRegex            = regexp.MustCompile(`(?s)<hiero.*?</hiero>`)
	HieroglyphRegex            = regexp.MustCompile(`(?s)<chem.*?</chem>`)
	SyntaxHighlightRegex = regexp.MustCompile(`(?s)<syntaxhighlight.*?</syntaxhighlight>`)
	PreRegex = regexp.MustCompile(`(?s)<pre.*?</pre>`)
	BrRegex              = regexp.MustCompile(`<(p|br).*?>`)

	AlteredQuote = regexp.MustCompile(`\[([a-zA-Z])]`)

	link = regexp.MustCompile(`\[http[^]]+]`)

	RemoveLinks = regexp.MustCompile(`\[\[(:?Category:|List of)[^]]+]]`)

	WikipediaLinks = regexp.MustCompile(`\[\[([^[\]]+\|)?([^[|]+?)]]`)

	widgets = regexp.MustCompile(`{[^{}]*}`)

	RefRegex = regexp.MustCompile(`(?s)<ref.*?(>.*?</ref>| ?/>)`)
)

func keepReplacing(pattern *regexp.Regexp, text string, replace string) string {
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
	text = keepReplacing(WikipediaLinks, text, "$2")

	text = RefRegex.ReplaceAllString(text, "")

	text = CommentRegex.ReplaceAllString(text, "")
	text = GalleryRegex.ReplaceAllString(text, "")
	text = BlockQuoteRegex.ReplaceAllString(text, "")
	text = SubSupRegex.ReplaceAllString(text, "")
	text = TimelineRegex.ReplaceAllString(text, "")
	text = MathRegex.ReplaceAllString(text, MathToken)
	text = HieroglyphRegex.ReplaceAllString(text, HieroglyphToken)
	text = CodeRegex.ReplaceAllString(text, "")
	text = ChemRegex.ReplaceAllString(text, "")
	text = ImageMapRegex.ReplaceAllString(text, "")
	text = SyntaxHighlightRegex.ReplaceAllString(text, "")
	text = PreRegex.ReplaceAllString(text, "")

	text = IgnoredTagsRegex.ReplaceAllString(text, "")

	text = BrRegex.ReplaceAllString(text, "\n")
	text = AlteredQuote.ReplaceAllString(text, "$1")

	lines := strings.Split(text, "\n")
	result := make([]string, 0, len(lines))
	lastLineEmpty := false

	skip := false

	for _, line := range lines {

		// line = WikipediaLinks.ReplaceAllString(line, "$1")

		line = strings.ReplaceAll(line, "&nbsp;", " ")
		line = strings.ReplaceAll(line, "&ndash;", "–")

		line = link.ReplaceAllString(line, "")

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
