package article

import (
	"regexp"
	"strings"
)

var (
	LinkStartPattern = regexp.MustCompile(`\[\[`)
	LinkEndPattern   = regexp.MustCompile(`]]`)
)

type LinkStart struct{}

func (t LinkStart) Render() string {
	return "[["
}

type LinkEnd struct{}

func (t LinkEnd) Render() string {
	return "]]"
}

type Link struct {
	Target  LiteralText
	Display LiteralText
}

func (t Link) Render() string {
	if t.Display != "" {
		return t.Display.Render()
	}
	return t.Target.Render()
}

func ParseLink(tokens []Token) Token {
	// Find first pipe
	text := Render(tokens[1:])
	splits := strings.SplitN(text, "|", 2)

	target := splits[0]
	target = strings.TrimSpace(target)
	result := Link{Target: LiteralText(target)}

	if len(splits) > 1 {
		display := splits[1]
		display = strings.TrimSpace(display)
		result.Display = LiteralText(display)
	}

	return result
}
