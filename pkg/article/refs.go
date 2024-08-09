package article

import "regexp"

var (
	RefStartPattern     = regexp.MustCompile(`<ref( name *= *("[^"]+"|[^">]+))? *>`)
	RefEndPattern       = regexp.MustCompile("</ref>")
	RefAutoClosePattern = regexp.MustCompile(`<ref name *= *("[^"]+"|[^">]+) */>`)
)

type RefStart string

func (t RefStart) Render() string {
	return string(t)
}

func ParseRefStart(s string) Token {
	return RefStart(s)
}

type RefEnd struct{}

func (t RefEnd) Render() string {
	return "</ref>"
}

func ParseRefEnd(string) Token {
	return RefEnd{}
}

type RefAutoClose string

func (t RefAutoClose) Render() string {
	return ""
}

func ParseRefAutoClose(s string) Token {
	return RefAutoClose(s)
}

type Ref struct {
	Tokens []Token
}

func (t Ref) Render() string {
	return ""
}

func (t RefEnd) Backtrack(tokens []Token) (Token, int) {
	_, startIdx, found := BacktrackUntil[RefStart](tokens)
	if !found {
		return nil, startIdx
	}

	return Ref{Tokens: tokens[startIdx:]}, startIdx
}
