package article

import (
	"regexp"
	"strings"
)

// allowedTags are the HTML tags permitted in wikitext.
// List is in the same order as on https://en.wikipedia.org/wiki/Help:HTML_in_wikitext.
var allowedTags = []string{
	"h1",
	"h2",
	"h3",
	"h4",
	"h5",
	"h6",
	"p",
	"abbr",
	"b",
	"bdi",
	"bdo",
	"blockquote",
	"cite",
	"code",
	"data",
	"del",
	"dfn",
	"em",
	"i",
	"ins",
	"kbd",
	"mark",
	"pre",
	"q",
	"rp",
	"rt",
	"ruby",
	"s",
	"samp",
	"small",
	"strong",
	"sub",
	"sup",
	"time",
	"u",
	"var",
	"wbr",
	"dl",
	"dt",
	"dd",
	"ol",
	"ul",
	"li",
	"div",
	"span",
	"table",
	"td",
	"tr",
	"th",
	"caption",
	"thead",
	"tfoot",
	"tbody",
}

var (
	HTMLOpenTagPattern  = regexp.MustCompile("<(" + strings.Join(allowedTags, "|") + ")>")
	HTMLCloseTagPattern = regexp.MustCompile("</(" + strings.Join(allowedTags, "|") + ")>")
)

type HTMLOpenTag string

func (t HTMLOpenTag) Render() string {
	return string(t)
}

func (t HTMLOpenTag) Name() string {
	spaceIdx := strings.IndexRune(string(t), ' ')

	if spaceIdx == -1 {
		return string(t[1 : len(t)-1])
	}

	return string(t[1:spaceIdx])
}

func ParseHTMLOpenTag(s string) Token {
	return HTMLOpenTag(s)
}

type HTMLCloseTag string

func (t HTMLCloseTag) Render() string {
	return string(t)
}

func ParseHTMLCloseTag(s string) Token {
	return HTMLCloseTag(s)
}

func (t HTMLCloseTag) Name() string {
	return string(t[2 : len(t)-1])
}

var standaloneTags = []string{
	"br",
	"hr",
}

func (t HTMLCloseTag) Backtrack(tokens []Token) (Token, int) {
	wantName := t.Name()

	startIdx := len(tokens) - 1
	for ; startIdx >= 0; startIdx-- {
		start, ok := tokens[startIdx].(HTMLOpenTag)
		if !ok {
			continue
		}

		if start.Name() == wantName {
			return ParseHTMLTag(start, tokens[startIdx:]), startIdx
		}
	}

	return nil, -1
}

type NilToken struct{}

func (t NilToken) Render() string {
	return "{nil}"
}

func ParseHTMLTag(start HTMLOpenTag, tokens []Token) Token {
	switch start.Name() {
	case "blockquote":
		return ParseBlockquote(tokens)
	case "em":
		return ParseEmphasis(tokens)
	case "math":
		return ParseMath(tokens)
	case "sub":
		return ParseSubscript(tokens)
	case "sup":
		return ParseSuperscript(tokens)
	}

	return NilToken{}
}
