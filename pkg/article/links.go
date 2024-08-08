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

func (t LinkStart) Original() string {
	return "[["
}

func ParseLinkStart(string) Token {
	return LinkStart{}
}

type LinkEnd struct{}

func (t LinkEnd) Original() string {
	return "]]"
}

func ParseLinkEnd(string) Token {
	return LinkEnd{}
}

type Link struct {
	Target  LiteralText
	Display LiteralText
}

func (t Link) Original() string {
	if t.Display != "" {
		return t.Display.Original()
	}
	return t.Target.Original()
}

type LinkFile struct {
	Target  LiteralText
	Caption []Token
}

func (t LinkFile) Original() string {
	return Render(t.Caption)
}

func (t LinkEnd) Backtrack(tokens []Token) (Token, int) {
	_, startIdx, found := BacktrackUntil[LinkStart](tokens)
	if !found {
		return nil, startIdx
	}

	return ParseLink(tokens[startIdx:]), startIdx
}

func ParseLink(tokens []Token) Token {
	// Find first pipe
	text := Render(tokens[1:])
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

	return result
}

func ParseLinkFile(target string, args string) Token {
	splits := strings.Split(args, "|")

	caption := Tokenize(UnparsedText(splits[len(splits)-1]))

	return LinkFile{
		Target:  LiteralText(target),
		Caption: caption,
	}
}
