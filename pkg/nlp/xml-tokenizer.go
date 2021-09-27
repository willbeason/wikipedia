package nlp

import (
	"regexp"
)

// XMLTagRegex tries to find XML tags which are still present in the corpus. Useful for finding
// problematic tags that we want to avoid.
var XMLTagRegex = regexp.MustCompile(`<[a-z][a-z0-9]+`)

type XMLTokenizer struct{}

var _ Tokenizer = XMLTokenizer{}

func (x XMLTokenizer) Tokenize(s string) []string {
	return XMLTagRegex.FindAllString(s, -1)
}
