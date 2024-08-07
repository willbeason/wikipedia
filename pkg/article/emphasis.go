package article

import "regexp"

var (
	EmphasisStartPattern = regexp.MustCompile(`<em>`)
	EmphasisEndPattern   = regexp.MustCompile(`</em>`)
)

type EmphasisStart struct{}

func (t EmphasisStart) Render() string {
	return "<em>"
}

func ParseEmphasisStart(string) Token {
	return EmphasisStart{}
}

type EmphasisEnd struct{}

func (t EmphasisEnd) Render() string {
	return "</em>"
}

func ParseEmphasisEnd(string) Token {
	return EmphasisEnd{}
}

type Emphasis struct {
	Quote []Token
}

func (t Emphasis) Render() string {
	return Render(t.Quote)
}

func ParseEmphasis(tokens []Token) (Token, error) {
	return Emphasis{tokens[1 : len(tokens)-1]}, nil
}
