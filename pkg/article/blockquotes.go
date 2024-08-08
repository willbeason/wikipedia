package article

import (
	"regexp"
)

var (
	BlockquoteStartPattern = regexp.MustCompile(`<blockquote>`)
	BlockquoteEndPattern   = regexp.MustCompile(`</blockquote>`)
)

type BlockquoteStart struct{}

func (t BlockquoteStart) Original() string {
	return "<blockquote>"
}

func ParseBlockquoteStart(string) Token {
	return BlockquoteStart{}
}

type BlockquoteEnd struct{}

func (t BlockquoteEnd) Original() string {
	return "</blockquote>"
}

func ParseBlockquoteEnd(string) Token {
	return BlockquoteEnd{}
}

type Blockquote struct {
	Quote []Token
}

func (t Blockquote) Original() string {
	return "\n" + Render(t.Quote) + "\n"
}

func ParseBlockquote(tokens []Token) Token {
	quote := Tokenize(UnparsedText(Render(tokens[1 : len(tokens)-1])))

	return Blockquote{Quote: quote}
}
