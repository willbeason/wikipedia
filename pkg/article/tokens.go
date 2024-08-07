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
