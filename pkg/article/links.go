package article

import (
	"fmt"
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

func (t LinkEnd) Backtrack(tokens []Token) (Token, int) {
	_, startIdx, found := BacktrackUntil[LinkStart](tokens)
	if !found {
		return nil, startIdx
	}

	if startIdx == len(tokens)-1 {
		// This is an empty link, literally just "[[]]".
		return LiteralText("[[]]"), startIdx
	}

	return ParseLink(tokens[startIdx:]), startIdx
}

func ParseLink(tokens []Token) Token {
	// Find first pipe
	text := Render(tokens[1:])
	splits := strings.SplitN(text, "|", 2)

	target := splits[0]
	target = strings.TrimSpace(target)
	if len(target) == 0 {
		return LiteralText(fmt.Sprintf("[[%s]]", splits[0]))
	}

	if strings.HasPrefix(target, "File:") {
		return ParseLinkFile(target, splits[1:]...)
	}

	result := Link{Target: LiteralText(target)}

	if len(splits) > 1 {
		display := splits[1]
		display = strings.TrimSpace(display)
		result.Display = LiteralText(display)
	}

	return result
}

func ParseLinkFile(target string, args ...string) Token {
	result := LinkFile{
		Target: LiteralText(target),
	}

	if len(args) > 0 {
		splits := strings.Split(args[0], "|")
		result.Caption = Tokenize(UnparsedText(splits[len(splits)-1]))
	}

	return result
}

func ToLinkTargets(tokens []Token, ignoredSections map[string]bool) []string {
	var result []string

	for _, token := range tokens {
		switch l := token.(type) {
		case Link:
			target := l.Target.Render()
			// if len(target) == 0 {
			//	fmt.Println(n, "/", len(tokens))
			//	for i, t := range tokens {
			//		fmt.Printf("%d. %T %q\n", i, t, t.Render())
			//	}
			//}
			target = strings.ToUpper(target[0:1]) + target[1:]
			if strings.Contains(target, "#") {
				target = strings.Split(target, "#")[0]
			}
			result = append(result, target)
		case Section:
			if ignoredSections[l.Header.Render()] {
				continue
			}
			result = append(result, ToLinkTargets(l.Text, ignoredSections)...)
		}
	}

	return result
}
