package article

import "regexp"

var (
	BlockquoteStartPattern = regexp.MustCompile(`<blockquote>`)
	BlockquoteEndPattern   = regexp.MustCompile(`</blockquote>`)
)

type BlockquoteStart struct{}

func (t BlockquoteStart) Render() string {
	return "<blockquote>"
}

func ParseBlockquoteStart(string) Token {
	return BlockquoteStart{}
}

type BlockquoteEnd struct{}

func (t BlockquoteEnd) Render() string {
	return "</blockquote>"
}

func ParseBlockquoteEnd(string) Token {
	return BlockquoteEnd{}
}

type Blockquote struct {
	Quote []Token
}

func (t Blockquote) Render() string {
	return Render(t.Quote)
}

func ParseBlockquote(tokens []Token) (Token, error) {
	return Blockquote{tokens[1 : len(tokens)-1]}, nil
}
