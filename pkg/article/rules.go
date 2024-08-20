package article

import (
	"regexp"
	"strings"
)

type RuleFn func([]Token) []Token

// ToLiterals converts all unparsed text to LiteralText.
func ToLiterals(tokens []Token) []Token {
	result := make([]Token, len(tokens))

	for i, token := range tokens {
		unparsed, isUnparsed := token.(UnparsedText)
		if isUnparsed {
			result[i] = LiteralText(unparsed)
		} else {
			result[i] = token
		}
	}

	return result
}

func ExactTokenRule(pattern string, newToken func() Token) RuleFn {
	return func(tokens []Token) []Token {
		var result []Token

		for _, token := range tokens {
			unparsed, isUnparsed := token.(UnparsedText)
			if !isUnparsed {
				result = append(result, token)
				continue
			}

			raw := string(unparsed)
			splits := strings.Split(raw, pattern)
			if len(splits[0]) > 0 {
				// Only append beginning as unparsed text if nonempty.
				result = append(result, UnparsedText(splits[0]))
			}

			for i := 1; i < len(splits); i++ {
				split := splits[i]
				result = append(result, newToken())
				if len(split) > 0 {
					result = append(result, UnparsedText(split))
				}
			}
		}

		return result
	}
}

// PatternTokenRule searches for matches for a regular expression in currently-unparsed token in a list of tokens.
//
// newToken creates new tokens for each match.
func PatternTokenRule(regex *regexp.Regexp, newToken func(string) Token) RuleFn {
	return func(tokens []Token) []Token {
		var result []Token

		for _, token := range tokens {
			unparsed, isUnparsed := token.(UnparsedText)
			if !isUnparsed {
				result = append(result, token)
				continue
			}

			matches := regex.FindAllStringIndex(string(unparsed), -1)
			lastEnd := 0
			for _, match := range matches {
				if lastEnd < match[0] {
					result = append(result, unparsed[lastEnd:match[0]])
				}
				result = append(result, newToken(string(unparsed[match[0]:match[1]])))
				lastEnd = match[1]
			}

			if lastEnd < len(unparsed) {
				// Append the entire token if there were no matches.
				result = append(result, unparsed[lastEnd:])
			}
		}

		return result
	}
}

// MergeTokenRule is a higher-order function that creates a rule function for merging tokens.
// The rule function takes a slice of tokens and returns a new slice of tokens with tokens between ranges of START and
// END tokens merged according to parseToken. This merging function parseToken, accepts tokens from START to END
// inclusive.
//
// Attempts to merge the deepest token sequences first. So [START, START, END, END] will result in [START, MERGED, END].
// Apply repeatedly to merge all instances.
//
// The MergeTokenRule function takes two type parameters, START and END, that represent the tokens marking the beginning
// and ending of sequences of tokens to merge.
//
// Example usage:
//
//	MergeTokenRule[RefStart, RefEnd](ParseRef) - merges tokens between tokens of type RefStart and tokens of type
//	RefEnd using the ParseRef function.
func MergeTokenRule[START, END Token](
	parseToken func([]Token) Token,
) RuleFn {
	return func(tokens []Token) []Token {
		var result []Token

		lastStartIdx := -1

		for idx, token := range tokens {
			_, isStart := token.(START)
			if isStart {
				if lastStartIdx != -1 {
					// We only want to merge a START with an END if there are no intervening START tokens.
					// This ignores previously-seen START tokens; those will get merged in future applications of this
					// rule.
					result = append(result, tokens[lastStartIdx:idx]...)
				}

				lastStartIdx = idx
				continue
			}

			if lastStartIdx == -1 {
				result = append(result, token)
				continue
			}

			_, isEnd := token.(END)
			if !isEnd {
				continue
			}

			parsed := parseToken(tokens[lastStartIdx : idx+1])

			result = append(result, parsed)

			lastStartIdx = -1
		}

		// Since there is no END token to match the most recently-seen START token, append all of those.
		// This represents an error in the original wikitext we can't resolve.
		if lastStartIdx != -1 {
			result = append(result, tokens[lastStartIdx:]...)
		}

		return result
	}
}
