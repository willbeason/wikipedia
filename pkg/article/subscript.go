package article

type Subscript struct {
	Quote []Token
}

func (t Subscript) Render() string {
	return "_" + Render(t.Quote)
}

func ParseSubscript(tokens []Token) Token {
	return Subscript{Quote: append([]Token{}, tokens[1:]...)}
}
