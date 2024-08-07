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

func ParseLinkStart(string) Token {
	return LinkStart{}
}

type LinkEnd struct{}

func (t LinkEnd) Render() string {
	return "]]"
}

func ParseLinkEnd(string) Token {
	return LinkEnd{}
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

type LinkFile struct {
	Target  LiteralText
	Caption []Token
}

func (t LinkFile) Render() string {
	return Render(t.Caption)
}

func ParseLink(tokens []Token) (Token, error) {
	// Find first pipe
	text := Render(tokens[1 : len(tokens)-1])
	splits := strings.SplitN(text, "|", 2)

	target := splits[0]
	target = strings.TrimSpace(target)

	if strings.HasPrefix(target, "File:") {
		return ParseLinkFile(target, splits[1])
	}

	result := Link{Target: LiteralText(target)}

	if len(splits) > 1 {
		display := splits[1]
		display = strings.TrimSpace(display)
		result.Display = LiteralText(display)
	}

	return result, nil
}

func ParseLinkFile(target string, args string) (Token, error) {
	splits := strings.Split(args, "|")

	caption, err := Tokenize(UnparsedText(splits[len(splits)-1]))
	if err != nil {
		return nil, err
	}

	return LinkFile{
		Target:  LiteralText(target),
		Caption: caption,
	}, nil
}
