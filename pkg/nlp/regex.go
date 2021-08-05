package nlp

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	NumToken  = "_num_"
	DateToken = "_date_"

	Months = "(january|february|march|april|may|june|july|august|september|october|november|december)"
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
	DivRegex        = regexp.MustCompile(`</?div.*?>`)
	GalleryRegex    = regexp.MustCompile(`</?gallery.*?>`)
	SpanRegex       = regexp.MustCompile(`</?span.*?>`)
	BigRegex        = regexp.MustCompile(`</?big.*?>`)
	PoemRegex       = regexp.MustCompile(`</?poem.*?>`)
	BlockQuoteRegex = regexp.MustCompile(`</?sup.*?>`)
	SupRegex        = regexp.MustCompile(`</?blockquote.*?>`)
	TimelineRegex   = regexp.MustCompile(`(?s)<timeline.*?</timeline>`)
	BrRegex         = regexp.MustCompile(`<br.*?>`)

	NumberRegex = regexp.MustCompile(`\b\d+(,\d{3})*(\.\d+)?\b`)
	DateRegex   = regexp.MustCompile(fmt.Sprintf(`(?i)\b(%s %s,? %s|%s %s,? %s)\b`,
		NumToken, Months, NumToken,
		Months, NumToken, NumToken,
	))
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

	text = BrRegex.ReplaceAllString(text, "\n")

	// Tokens for special types of sequences.
	text = NumberRegex.ReplaceAllString(text, NumToken)
	text = DateRegex.ReplaceAllString(text, DateToken)

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
