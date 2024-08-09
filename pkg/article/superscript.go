package article

import "regexp"

var (
	SuperscriptStartPattern = regexp.MustCompile(`<sup>`)
	SuperscriptEndPattern   = regexp.MustCompile(`</sup>`)
)

type SuperscriptStart struct{}

func (t SuperscriptStart) Render() string {
	return "<sup>"
}

func ParseSuperscriptStart(string) Token {
	return SuperscriptStart{}
}

type SuperscriptEnd struct{}

func (t SuperscriptEnd) Render() string {
	return "</sup>"
}

func ParseSuperscriptEnd(string) Token {
	return SuperscriptEnd{}
}

type Superscript struct {
	Quote []Token
}

func (t Superscript) Render() string {
	return "^" + Render(t.Quote)
}

func (t SuperscriptEnd) Backtrack(tokens []Token) (Token, int) {
	_, startIdx, found := BacktrackUntil[SuperscriptStart](tokens)
	if !found {
		return nil, startIdx
	}

	return ParseSuperscript(tokens[startIdx:]), startIdx
}

func ParseSuperscript(tokens []Token) Token {
	return Superscript{Quote: append([]Token{}, tokens[1:]...)}
}
