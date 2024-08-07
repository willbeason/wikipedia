package article

import "regexp"

var (
	MathStartPattern = regexp.MustCompile(`<math>`)
	MathEndPattern   = regexp.MustCompile(`</math>`)
)

const MathToken = "_math_"

type MathStart struct{}

func (t MathStart) Render() string {
	return "<math>"
}

func ParseMathStart(string) Token {
	return MathStart{}
}

type MathEnd struct{}

func (t MathEnd) Render() string {
	return "</math>"
}

func ParseMathEnd(string) Token {
	return MathEnd{}
}

type Math struct {
	Quote []Token
}

func (t Math) Render() string {
	return MathToken
}

func ParseMath(tokens []Token) (Token, error) {
	return Math{tokens[1 : len(tokens)-1]}, nil
}
