package article

import "strings"

// Token represents a semantic unit of an article.
type Token interface {
	Render() string
}

type UnparsedText string

func (t UnparsedText) Render() string {
	panic("attempt to render unparsed text")
}

type LiteralText string

func (t LiteralText) Render() string {
	return string(t)
}

func Render(tokens []Token) string {
	sb := strings.Builder{}

	for _, t := range tokens {
		sb.WriteString(t.Render())
	}

	return sb.String()
}
