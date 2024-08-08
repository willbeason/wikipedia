package article

import "regexp"

var (
	MathStartPattern = regexp.MustCompile(`<math>`)
	MathEndPattern   = regexp.MustCompile(`</math>`)
)

const MathToken = "_math_"

type MathStart struct{}

func (t MathStart) Original() string {
	return "<math>"
}

func ParseMathStart(string) Token {
	return MathStart{}
}

type MathEnd struct{}

func (t MathEnd) Original() string {
	return "</math>"
}

func ParseMathEnd(string) Token {
	return MathEnd{}
}

type Math struct {
	Quote []Token
}

func (t Math) Original() string {
	return MathToken
}

func (t MathEnd) Backtrack(tokens []Token) (Token, int) {
	_, startIdx, found := BacktrackUntil[MathStart](tokens)
	if !found {
		return nil, startIdx
	}

	return ParseMath(tokens[startIdx:]), startIdx
}

func ParseMath(tokens []Token) Token {
	return Math{append([]Token{}, tokens[1:]...)}
}
