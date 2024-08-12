package article

type Superscript struct {
	Quote []Token
}

func (t Superscript) Render() string {
	return "^" + Render(t.Quote)
}

func ParseSuperscript(tokens []Token) Token {
	return Superscript{Quote: append([]Token{}, tokens[1:]...)}
}
