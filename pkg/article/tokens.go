package articles

// Token represents a semantic unit of an article.
type Token interface {
	Render() string
}

type UnparsedText string

func (t UnparsedText) Render() string {
	panic("attempt to render unparsed text")
}

type LiteralText string

func (t LiteralText) Render() string {
	return string(t)
}
