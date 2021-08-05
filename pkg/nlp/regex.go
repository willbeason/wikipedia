package nlp

import (
	"fmt"
	"regexp"
	"strings"
)

// Tokens to replace longer sequences with, that we treat as semantically identical for analysis.
const (
	NumToken  = "_num_"
	DateToken = "_date_"
	MathToken = "_math_"
	HieroglyphToken = "_hieroglyph_"
)

const (
	Months = "(january|february|march|april|may|june|july|august|september|october|november|december)"
)

var (
	WordRegex   = regexp.MustCompile(`[\w']+`)
	LetterRegex = regexp.MustCompile(`[A-Za-z]`)


	NumberRegex = regexp.MustCompile(`\b\d+(,\d{3})*(\.\d+)?\b`)
	DateRegex   = regexp.MustCompile(fmt.Sprintf(`\b(%s %s,? %s|%s %s,? %s)\b`,
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
	// For now, analyze articles in a case-insensitive manner.
	text = strings.ToLower(text)

	// Tokens for special types of sequences. For our current analyses we treat these as individual
	// identical "words".
	text = NumberRegex.ReplaceAllString(text, NumToken)
	text = DateRegex.ReplaceAllString(text, DateToken)

	return text
}

func IsArticle(title string) bool {
	return !strings.HasPrefix(title, "Wikipedia:") &&
		!strings.HasPrefix(title, "Category:") &&
		!strings.HasPrefix(title, "Draft:") &&
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
