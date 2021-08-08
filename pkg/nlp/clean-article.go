package nlp

import (
	"fmt"
	"regexp"
	"strings"
)

// ignoredTags are tags we can safely strip out, retaining the contents.
func ignoredTags() []string {
	return []string{
		"a",
		"abbr",
		"b",
		"bdi",
		"big",
		"blockquote",
		"c",
		"center",
		"div",
		"dfn",
		"dt",
		"em",
		"font",
		"h\\d",
		"http",
		"https",
		"i",
		"kbd",
		"li",
		"(no|only)?include(only)?",
		"ol",
		"mapframe",
		"nowiki",
		"q",
		"ref",
		"s",
		"small",
		"span",
		"su[bp]",
		"td",
		"tr",
		"time",
		"tt",
		"u",
		"var",
		"wbr",
		"www",
	}
}

// Regular expressions for cleaning Wikipedia articles of XML tags and formatting.
var (


	// CommentRegex matches commented-out text. Such text is not shown on pages
	// and is generally either off-topic or low quality.
	//
	// Obviously not perfect and can match non-comments in rare cases.
	CommentRegex = regexp.MustCompile("(?s)<!--.*?-->")

	IgnoredTagsRegex     = regexp.MustCompile(fmt.Sprintf(`(?i)</?(%s).*?>`, strings.Join(ignoredTags(), "|")))
	TimelineRegex        = regexp.MustCompile(`(?is)<timeline.*?</timeline[\w\s]*>`)
	GalleryRegex         = regexp.MustCompile(`(?is)<gallery.*?</gallery[\w\s]*>`)
	GraphRegex         = regexp.MustCompile(`(?is)<graph.*?</graph[\w\s]*>`)
	ImageMapRegex        = regexp.MustCompile(`(?is)<imagemap.*?</imagemap[\w\s]*>`)
	MathRegex            = regexp.MustCompile(`(?is)<math.*?</math[\w\s]*>`)
	CodeRegex            = regexp.MustCompile(`(?is)<code.*?</code[\w\s]*>`)
	CiteRegex            = regexp.MustCompile(`(?is)<cite.*?</cite[\w\s]*>`)
	ChemRegex            = regexp.MustCompile(`(?is)<chem.*?</chem[\w\s]*>`)
	PoemRegex            = regexp.MustCompile(`(?is)<poem.*?</poem[\w\s]*>`)
	HieroglyphRegex      = regexp.MustCompile(`(?is)<hiero.*?</hiero[\w\s]*>`)
	MapframeRegex      = regexp.MustCompile(`(?is)<mapframe.*?</mapframe[\w\s]*>`)
	DelRegex      = regexp.MustCompile(`(?is)<del.*?</del[\w\s]*>`)
	SyntaxHighlightRegex = regexp.MustCompile(`(?is)<syntaxhighlight.*?</syntaxhighlight[\w\s]*>`)
	PreRegex             = regexp.MustCompile(`(?is)<pre.*?</pre[\w\s]*>`)
	TableRegex             = regexp.MustCompile(`(?is)<table.*?</table[\w\s]*>`)
	BrRegex              = regexp.MustCompile(`(?i)<(p|br|hr).*?>`)

	AlteredQuote = regexp.MustCompile(`\[([a-zA-Z])]`)

	link = regexp.MustCompile(`\[http[^]]+]`)

	RemoveLinks = regexp.MustCompile(`\[\[(:?Category:|List of)[^]]+]]`)

	WikipediaLinks = regexp.MustCompile(`\[\[([^[\]]+\|)?([^[|]+?)]]`)

	widgets = regexp.MustCompile(`{[^{}]*}`)
	parens  = regexp.MustCompile(`\([^()]*?\)`)

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
	text = keepReplacing(parens, text, "")

	text = RefRegex.ReplaceAllString(text, "")

	text = CommentRegex.ReplaceAllString(text, "")
	text = TableRegex.ReplaceAllString(text, "")
	text = CiteRegex.ReplaceAllString(text, "")
	text = GalleryRegex.ReplaceAllString(text, "")
	text = GraphRegex.ReplaceAllString(text, "")
	text = TimelineRegex.ReplaceAllString(text, "")
	text = MathRegex.ReplaceAllString(text, MathToken)
	text = HieroglyphRegex.ReplaceAllString(text, HieroglyphToken)
	text = CodeRegex.ReplaceAllString(text, "")
	text = ChemRegex.ReplaceAllString(text, "")
	text = ImageMapRegex.ReplaceAllString(text, "")
	text = SyntaxHighlightRegex.ReplaceAllString(text, "")
	text = PreRegex.ReplaceAllString(text, "")
	text = PoemRegex.ReplaceAllString(text, "")
	text = DelRegex.ReplaceAllString(text, "")
	text = MapframeRegex.ReplaceAllString(text, "")

	text = BrRegex.ReplaceAllString(text, "\n")
	text = AlteredQuote.ReplaceAllString(text, "$1")

	text = IgnoredTagsRegex.ReplaceAllString(text, "")

	lines := strings.Split(text, "\n")
	result := make([]string, 0, len(lines))
	lastLineEmpty := false

	skip := false

	for _, line := range lines {
		line = strings.ReplaceAll(line, "&nbsp;", " ")
		line = strings.ReplaceAll(line, "&ndash;", "–")

		line = strings.TrimPrefix(line, "*")
		line = strings.TrimSpace(line)

		switch strings.ToLower(strings.Trim(line, " =")) {
		case "bibliography", "citations", "external links", "further reading", "notes", "references", "see also", "sources":
			skip = true
		default:
			if strings.HasPrefix(line, "==") {
				skip = false
			}
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
