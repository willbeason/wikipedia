package article

const MathToken = "_math_"

type Math struct {
	Quote []Token
}

func (t Math) Render() string {
	return MathToken
}

func ParseMath(tokens []Token) Token {
	return Math{append([]Token{}, tokens[1:]...)}
}
