package article

import "strings"

// Token represents a semantic unit of an article.
type Token interface {
	// Original returns the exact original text which was parsed into this Token.
	Original() string
}

type EndToken interface {
	Token
	// MatchesStart returns true if the EndToken should be merged with the passed Token.
	MatchesStart(t Token) bool
	// Merge merges a sequence of Tokens into one, assuming this is the end of the sequence.
	Merge(tokens []Token) Token
}

type UnparsedText string

func (t UnparsedText) Original() string {
	panic("attempt to render unparsed text")
}

type LiteralText string

const NBSP = `&nbsp;`

func (t LiteralText) Original() string {
	s := string(t)
	s = strings.ReplaceAll(s, NBSP, " ")
	return s
}

func Render(tokens []Token) string {
	sb := strings.Builder{}

	for _, t := range tokens {
		sb.WriteString(t.Original())
	}

	return sb.String()
}
