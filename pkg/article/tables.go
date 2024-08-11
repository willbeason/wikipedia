package article

var (
	TableStartPattern = `{|`
	TableEndPattern   = `|}`
)

type TableStart string

func (t TableStart) Render() string {
	return string(t)
}

func ParseTableStart(s string) Token {
	return TableStart(s)
}

type TableEnd string

func (t TableEnd) Render() string {
	return string(t)
}

func ParseTableEnd(s string) Token {
	return TableEnd(s)
}

type Table struct {
	Quote []Token
}

func (t Table) Render() string {
	return ""
}

func (t TableEnd) Backtrack(tokens []Token) (Token, int) {
	_, startIdx, found := BacktrackUntil[TableStart](tokens)
	if !found {
		return nil, startIdx
	}

	return ParseTable(tokens[startIdx:]), startIdx
}

func ParseTable(tokens []Token) Token {
	return Table{Quote: append([]Token{}, tokens[1:]...)}
}
