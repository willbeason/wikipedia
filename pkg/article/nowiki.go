package article

import (
	"regexp"
	"strings"
)

const (
	NowikiSectionStartPattern = `<nowiki>`
	NowikiSectionEndPattern   = `</nowiki>`
)

var NowikiAutoClosePattern = regexp.MustCompile(`<nowiki ?/>`)

type NowikiStart struct{}

func (t NowikiStart) Original() string {
	return NowikiSectionStartPattern
}

type NowikiEnd struct{}

func (t NowikiEnd) Original() string {
	return NowikiSectionEndPattern
}

type NowikiAutoClose struct{}

func (t NowikiAutoClose) Original() string {
	return ""
}

func ParseNowikiAutoClose(string) Token {
	return NowikiAutoClose{}
}

// Nowiki represents a section of text which should be displayed as-is.
type Nowiki string

func (t Nowiki) Original() string {
	return string(t)
}

func (t NowikiEnd) Merge(tokens []Token) Token {
	sb := strings.Builder{}
	for _, token := range tokens {
		switch s := token.(type) {
		case UnparsedText:
			// Don't need to parse text in Nowiki sections.
			sb.WriteString(string(s))
		default:
			// Get the original text of any parsed tokens.
			sb.WriteString(s.Original())
		}
	}

	return Nowiki(sb.String())
}

// MergeNowikiTokens merges <nowiki> sections into literal text to avoid further parsing.
//
// <nowiki> sections cannot be nested or overlap, and so the first <nowiki> token is always closed by the first
// </nowiki> token.
//
// Further, unclosed <nowiki> sections are parsed as normal and are not an error.
//
// The boolean value indicates whether any tokens were merged.
func MergeNowikiTokens(tokens []Token) []Token {
	var result []Token

	var queue []Token
	for _, token := range tokens {
		switch t := token.(type) {
		case NowikiStart:
			// Either opens a new nowiki environment or appends as a token that will be rendered as-is.
			// We don't care about nested nowiki environments - only the first-started one wins.
			queue = append(queue, t)
		case NowikiEnd:
			if len(queue) == 0 {
				// Close token with no corresponding open, so must be literal text.
				result = append(result, t)
			} else {
				// Original tokens in the queue, excluding the initial token which is not rendered.
				result = append(result, t.Merge(queue[1:]))

				// We don't care about subsequent close tokens until we see an open token.
				queue = nil
			}
		default:
			if len(queue) == 0 {
				// We are in the normal environment, so append as normal.
				result = append(result, token)
			} else {
				// We are in the <nowiki> environment, so append the token.
				queue = append(queue, token)
			}
		}
	}

	// Append any unresolved open nowiki sections.
	result = append(result, queue...)

	return result
}
