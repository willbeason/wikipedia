package article

import (
	"regexp"
	"strings"
)

var ExternalLinkPattern = regexp.MustCompile(`\[https?://[^ \]]+ [^]]+]`)

type ExternalLink struct {
	Target  string
	Display string
}

func ParseExternalLink(s string) Token {
	// Trim left and right square brackets.
	s = s[1 : len(s)-1]

	splits := strings.SplitN(s, " ", 2)

	return ExternalLink{
		Target:  splits[0],
		Display: splits[1],
	}
}

func (t ExternalLink) Render() string {
	return t.Display
}
