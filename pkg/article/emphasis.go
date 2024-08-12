package article

type Emphasis struct {
	Quote []Token
}

func (t Emphasis) Render() string {
	// fmt.Println(t.Quote[0].Render())
	return Render(t.Quote)
}

func ParseEmphasis(tokens []Token) Token {
	return Emphasis{Quote: append([]Token{}, tokens[1:]...)}
}
