package article

import "regexp"

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
	return Render(t.Quote)
}

func ParseEmphasis(tokens []Token) Token {
	return Emphasis{tokens[1 : len(tokens)-1]}
}
