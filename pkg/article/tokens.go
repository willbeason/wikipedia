package article

import "strings"

// Token represents a semantic unit of an article.
type Token interface {
	// Render returns the text as displayed of this Token.
	Render() string
}

type EndToken interface {
	Token
	// Backtrack looks through merges a sequence of Tokens into one, assuming this is the end of the sequence.
	Backtrack(tokens []Token) (Token, int)
}

type UnparsedText string

func (t UnparsedText) Render() string {
	panic("attempt to render unparsed text")
}

type LiteralText string

const NBSP = `&nbsp;`

func (t LiteralText) Render() string {
	s := string(t)
	s = strings.ReplaceAll(s, NBSP, " ")
	return s
}

func Render(tokens []Token) string {
	sb := strings.Builder{}

	for _, t := range tokens {
		sb.WriteString(t.Render())
	}

	return sb.String()
}

func BacktrackUntil[T Token](tokens []Token) (T, int, bool) {
	startIdx := len(tokens) - 1
	for startIdx >= 0 {
		if token, ok := tokens[startIdx].(T); ok {
			return token, startIdx, true
		}
		startIdx--
	}

	var zero T
	return zero, -1, false
}
