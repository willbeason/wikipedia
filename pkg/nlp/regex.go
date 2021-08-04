package nlp

import (
	"regexp"
	"strings"
)

var (
	WordRegex   = regexp.MustCompile(`[\w']+`)
	LetterRegex = regexp.MustCompile(`[A-Za-z]`)

	// CommentRegex matches commented-out text. Such text is not shown on pages
	// and is generally either off-topic or low quality.
	//
	// Obviously not perfect and can match non-comments in rare cases.
	CommentRegex = regexp.MustCompile("(?s)<!--.+?-->")

	CellRegex       = regexp.MustCompile(`{?[\|!].+`)
	DivRegex        = regexp.MustCompile(`(?s)</?div.*?>`)
	GalleryRegex    = regexp.MustCompile(`(?s)</?gallery.*?>`)
	SpanRegex       = regexp.MustCompile(`(?s)</?span.*?>`)
	BigRegex        = regexp.MustCompile(`(?s)</?big.*?>`)
	PoemRegex       = regexp.MustCompile(`(?s)</?poem.*?>`)
	BlockQuoteRegex = regexp.MustCompile(`(?s)</?sup.*?>`)
	SupRegex        = regexp.MustCompile(`(?s)</?blockquote.*?>`)
	TimelineRegex   = regexp.MustCompile(`(?s)<timeline.*?</timeline>`)

	NumberRegex = regexp.MustCompile(`\b\d+\b`)
)

func Normalize(w string) string {
	w = strings.ToLower(w)
	w = strings.Trim(w, "'")

	return w
}

func NormalizeArticle(text string) string {
	text = CommentRegex.ReplaceAllString(text, "")
	text = DivRegex.ReplaceAllString(text, "")
	text = CellRegex.ReplaceAllString(text, "")
	text = GalleryRegex.ReplaceAllString(text, "")
	text = SpanRegex.ReplaceAllString(text, "")
	text = BigRegex.ReplaceAllString(text, "")
	text = PoemRegex.ReplaceAllString(text, "")
	text = BlockQuoteRegex.ReplaceAllString(text, "")
	text = SupRegex.ReplaceAllString(text, "")
	text = TimelineRegex.ReplaceAllString(text, "")

	// Tokens for special types of sequences.
	text = NumberRegex.ReplaceAllString(text, "_number_")

	return text
}

func IsArticle(title string) bool {
	return !strings.HasPrefix(title, "Wikipedia:") &&
		!strings.HasPrefix(title, "Category:") &&
		!strings.HasPrefix(title, "Template:") &&
		!strings.HasPrefix(title, "File:") &&
		!strings.HasPrefix(title, "Portal:") &&
		!strings.HasPrefix(title, "Help:") &&
		!strings.HasPrefix(title, "List of ") &&
		!strings.HasPrefix(title, "Module:") &&
		!strings.HasPrefix(title, "MediaWiki:")
}

func HasLetter(word string) bool {
	return LetterRegex.MatchString(word)
}
