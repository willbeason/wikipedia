package article

import (
	"regexp"
)

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
	return "\n" + Render(t.Quote) + "\n"
}

func (t BlockquoteEnd) Backtrack(tokens []Token) (Token, int) {
	_, startIdx, found := BacktrackUntil[BlockquoteStart](tokens)
	if !found {
		return nil, startIdx
	}

	return ParseBlockquote(tokens[startIdx:]), startIdx
}

func ParseBlockquote(tokens []Token) Token {
	quote := Tokenize(UnparsedText(Render(tokens[1:])))

	return Blockquote{Quote: quote}
}
