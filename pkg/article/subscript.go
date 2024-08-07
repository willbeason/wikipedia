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

func ParseSubscript(tokens []Token) (Token, error) {
	return Subscript{tokens[1 : len(tokens)-1]}, nil
}
