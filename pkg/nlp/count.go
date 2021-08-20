package nlp

import "strings"

type Tokenizer interface {
	// Tokenize splits s into distinct tokens.
	Tokenize(s string) []string
}

type Counter struct {
	Tokenizer
}

func (c Counter) Count(s string) map[string]int {
	result := make(map[string]int)

	for _, word := range c.Tokenize(s) {
		result[word]++
	}

	return result
}

type WordTokenizer struct{}

func (t WordTokenizer) Tokenize(s string) []string {
	words := WordRegex.FindAllString(s, -1)
	tokens := make([]string, len(words))

	nTokens := 0
	for _, token := range words {
		normalizedToken := strings.Trim(token, "'")
		if normalizedToken == "" {
			// Ignore "words" which are just apostrophes or are empty string.
			continue
		}

		tokens[nTokens] = normalizedToken
		nTokens++
	}

	return tokens[:nTokens]
}

type NgramTokenizer struct {
	Underlying WordTokenizer

	Dictionary map[string]bool
}

func (t NgramTokenizer) Tokenize(s string) []string {
	if s == "" {
		return nil
	}

	words := t.Underlying.Tokenize(s)

	curStart := 0
	nTokens := 0
	nWords := len(words)
	tokens := make([]string, nWords)

	for curStart < nWords {
		curString := words[curStart]

		curLen := 1
		for ; curLen < nWords-curStart; curLen++ {
			nextString := curString + " " + words[curStart+curLen]

			if !t.Dictionary[nextString] {
				break
			}

			curString = nextString
		}

		tokens[nTokens] = curString
		nTokens++

		curStart += curLen
	}

	return tokens[:nTokens]
}
