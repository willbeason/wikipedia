package article

import (
	"fmt"
	"regexp"
)

var (
	EmphasisStartPattern = regexp.MustCompile(`<em>`)
	EmphasisEndPattern   = regexp.MustCompile(`</em>`)
)

type EmphasisStart struct{}

func (t EmphasisStart) Original() string {
	return "<em>"
}

func ParseEmphasisStart(string) Token {
	return EmphasisStart{}
}

type EmphasisEnd struct{}

func (t EmphasisEnd) Original() string {
	return "</em>"
}

func ParseEmphasisEnd(string) Token {
	return EmphasisEnd{}
}

type Emphasis struct {
	Quote []Token
}

func (t Emphasis) Original() string {
	fmt.Println(t.Quote[0].Original())
	return Render(t.Quote)
}

func (t EmphasisEnd) Backtrack(tokens []Token) (Token, int) {
	_, startIdx, found := BacktrackUntil[EmphasisStart](tokens)
	if !found {
		return nil, startIdx
	}

	return ParseEmphasis(tokens[startIdx:]), startIdx
}

func ParseEmphasis(tokens []Token) Token {
	return Emphasis{Quote: append([]Token{}, tokens[1:]...)}
}
