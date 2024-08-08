package article

import "regexp"

var (
	NowikiSectionStartPattern = regexp.MustCompile(`<nowiki>`)
	NowikiSectionEndPattern   = regexp.MustCompile(`</nowiki>`)
	NowikiAutoClosePattern    = regexp.MustCompile(`<nowiki ?/>`)
)

type NowikiStart struct{}

func (t NowikiStart) Render() string {
	return ""
}

type NowikiEnd struct{}

func (t NowikiEnd) Render() string {
	return ""
}

type NowikiAutoClose struct{}

func (t NowikiAutoClose) Render() string {
	return ""
}

func ParseNowikiAutoClose(string) Token {
	return NowikiAutoClose{}
}

func NowikiSectionTokens(tokens []Token) ([]Token, bool, error) {
	result := make([]Token, 0, len(tokens))
	appliedRule := false

	for _, token := range tokens {
		unparsed, isUnparsed := token.(UnparsedText)
		if !isUnparsed {
			result = append(result, token)
			continue
		}

		start := NowikiSectionStartPattern.FindStringIndex(string(unparsed))
		end := NowikiSectionEndPattern.FindStringIndex(string(unparsed))

		if start == nil || end == nil {
			// No matching <nowiki> tags.
			result = append(result, token)
			continue
		}

		appliedRule = true
		if start[0] > 0 {
			result = append(result, unparsed[:start[0]])
		}
		result = append(result, NowikiStart{}, LiteralText(unparsed[start[1]:end[0]]), NowikiEnd{})
		if end[1] < len(unparsed) {
			result = append(result, unparsed[end[1]:])
		}
	}

	return result, appliedRule, nil
}
