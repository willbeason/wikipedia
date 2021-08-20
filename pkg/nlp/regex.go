package nlp

import (
	"fmt"
	"regexp"
	"strings"
)

// Tokens to replace longer sequences with, that we treat as semantically identical for analysis.
const (
	NumToken        = "_num_"
	PercentToken    = "_percent_"
	DateToken       = "_date_"
	MathToken       = "_math_"
	HieroglyphToken = "_hieroglyph_" // nolint:gosec // This is a reference to egyptian hieroglyphs.
)

const (
	// Months are all the months of the year.
	Months = "(january|february|march|april|may|june|july|august|september|october|november|december)"
)

// Regular expressions for detecting semantically-similar sequences.
var (
	WordRegex   = regexp.MustCompile(`[\w']+`)
	LetterRegex = regexp.MustCompile(`[A-Za-z]`)

	NumberRegex  = regexp.MustCompile(`\b\d+(,\d{3})*(\.\d+)?\b`)
	PercentRegex = regexp.MustCompile(fmt.Sprintf(`%s%%`, NumToken))
	DateRegex    = regexp.MustCompile(fmt.Sprintf(`(?i)\b(%s (%s,? )?%s|%s %s,? %s)\b`,
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

	// Tokens for special types of sequences. For our current analyzes we treat these as individual
	// identical "words".
	text = NumberRegex.ReplaceAllString(text, NumToken)
	text = PercentRegex.ReplaceAllString(text, PercentToken)
	text = DateRegex.ReplaceAllString(text, DateToken)

	return text
}

func IsArticle(title string) bool {
	// Check in descending order of frequency.
	//
	// Observed 2021-08-01:
	//
	// Category	2,100,543
	// Wikipedia	1,170,425
	// File	915,410
	// Template	592,437
	// Portal	93,146
	// Draft	59,584
	// Module	12,598
	// MediaWiki	2,212
	// TimedText	1,352
	// Help	957
	return !strings.HasPrefix(title, "Category:") &&
		!strings.HasPrefix(title, "Wikipedia:") &&
		!strings.HasPrefix(title, "File:") &&
		!strings.HasPrefix(title, "Template:") &&
		!strings.HasPrefix(title, "Portal:") &&
		!strings.HasPrefix(title, "Draft:") &&
		!strings.HasPrefix(title, "Module:") &&
		!strings.HasPrefix(title, "MediaWiki:") &&
		!strings.HasPrefix(title, "TimedText:") &&
		!strings.HasPrefix(title, "Help:")
}

func HasLetter(word string) bool {
	return LetterRegex.MatchString(word)
}
