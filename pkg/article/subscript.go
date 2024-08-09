package article

import "regexp"

var (
	SubscriptStartPattern = regexp.MustCompile(`<sub>`)
	SubscriptEndPattern   = regexp.MustCompile(`</sub>`)
)

type SubscriptStart struct{}

func (t SubscriptStart) Render() string {
	return "<sub>"
}

func ParseSubscriptStart(string) Token {
	return SubscriptStart{}
}

type SubscriptEnd struct{}

func (t SubscriptEnd) Render() string {
	return "</sub>"
}

func ParseSubscriptEnd(string) Token {
	return SubscriptEnd{}
}

type Subscript struct {
	Quote []Token
}

func (t Subscript) Render() string {
	return "_" + Render(t.Quote)
}

func (t SubscriptEnd) Backtrack(tokens []Token) (Token, int) {
	_, startIdx, found := BacktrackUntil[SubscriptStart](tokens)
	if !found {
		return nil, startIdx
	}

	return ParseSubscript(tokens[startIdx:]), startIdx
}

func ParseSubscript(tokens []Token) Token {
	return Subscript{Quote: append([]Token{}, tokens[1:]...)}
}
