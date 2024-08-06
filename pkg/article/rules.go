package articles

import "regexp"

type RuleFn func([]Token) ([]Token, bool, error)

func ToLiterals(tokens []Token) ([]Token, bool, error) {
	var result []Token
	appliedRule := false

	for _, token := range tokens {
		unparsed, isUnparsed := token.(UnparsedText)
		if !isUnparsed {
			result = append(result, token)
			continue
		}

		appliedRule = true
		result = append(result, LiteralText(unparsed))
	}

	return result, appliedRule, nil
}

func PatternTokenRule(regex *regexp.Regexp, newToken func(string) Token) RuleFn {
	return func(tokens []Token) ([]Token, bool, error) {
		var result []Token
		appliedRule := false

		for _, token := range tokens {
			unparsed, isUnparsed := token.(UnparsedText)
			if !isUnparsed {
				result = append(result, token)
				continue
			}

			matches := regex.FindAllStringIndex(string(unparsed), -1)
			lastEnd := 0
			for _, match := range matches {
				appliedRule = true

				if lastEnd < match[0] {
					result = append(result, unparsed[lastEnd:match[0]])
				}
				result = append(result, newToken(string(unparsed[match[0]:match[1]])))
				lastEnd = match[1]
			}
			if lastEnd < len(unparsed) {
				result = append(result, unparsed[lastEnd:])
			}
		}

		return result, appliedRule, nil
	}
}
