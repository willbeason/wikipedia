package article

type Blockquote struct {
	Quote []Token
}

func (t Blockquote) Render() string {
	return "\n" + Render(t.Quote) + "\n"
}

func ParseBlockquote(tokens []Token) Token {
	quote := Tokenize(UnparsedText(Render(tokens[1:])))

	return Blockquote{Quote: quote}
}
