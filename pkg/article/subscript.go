package article

import "regexp"

var (
	SubscriptStartPattern = regexp.MustCompile(`<sub>`)
	SubscriptEndPattern   = regexp.MustCompile(`</sub>`)
)

type SubscriptStart struct{}

func (t SubscriptStart) Original() string {
	return "<sub>"
}

func ParseSubscriptStart(string) Token {
	return SubscriptStart{}
}

type SubscriptEnd struct{}

func (t SubscriptEnd) Original() string {
	return "</sub>"
}

func ParseSubscriptEnd(string) Token {
	return SubscriptEnd{}
}

type Subscript struct {
	Quote []Token
}

func (t Subscript) Original() string {
	return "_" + Render(t.Quote)
}

func ParseSubscript(tokens []Token) Token {
	return Subscript{tokens[1 : len(tokens)-1]}
}
