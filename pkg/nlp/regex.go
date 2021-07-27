package nlp

import (
	"regexp"
	"strings"
)

var (
	WordRegex = regexp.MustCompile(`[\w']+`)
	NumRegex = regexp.MustCompile(`^[0-9.]+$`)
	LetterRegex = regexp.MustCompile(`[A-Za-z]`)
)

func Normalize(w string) string {
	w = strings.ToLower(w)
	w = strings.Trim(w, "'")
	return w
}

func IsArticle(title string) bool {
	if strings.HasPrefix(title, "Wikipedia:") {
		return false
	}
	if strings.HasPrefix(title, "Category:") {
		return false
	}
	if strings.HasPrefix(title, "Template:") {
		return false
	}
	return !strings.HasPrefix(title, "File:")
}

func HasLetter(word string) bool {
	return LetterRegex.MatchString(word)
}
