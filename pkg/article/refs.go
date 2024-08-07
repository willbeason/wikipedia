package article

import "regexp"

var (
	RefStartPattern     = regexp.MustCompile(`<ref( name="[^"]+")? *>`)
	RefEndPattern       = regexp.MustCompile("</ref>")
	RefAutoclosePattern = regexp.MustCompile(`<ref name="[^"]+" */>`)
)

type RefStart string

func (t RefStart) Render() string {
	return string(t)
}

type RefEnd struct{}

func (t RefEnd) Render() string {
	return "</ref>"
}

type RefAutoclose string

func (t RefAutoclose) Render() string {
	return ""
}

type Ref struct {
	Tokens []Token
}

func (t Ref) Render() string {
	return ""
}

func ParseRef(tokens []Token) Token {
	return Ref{Tokens: tokens[1 : len(tokens)-1]}
}
