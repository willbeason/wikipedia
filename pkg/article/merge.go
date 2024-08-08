package article

func MergeTokens(tokens []Token) []Token {
	var result []Token

	for _, token := range tokens {
		endToken, isEnd := token.(EndToken)
		if !isEnd {
			// This isn't the end of a sequence of tokens.
			result = append(result, token)
			continue
		}

		for idx := len(result) - 1; idx >= 0; idx-- {
			if endToken.MatchesStart(result[idx]) {
				mergedToken := endToken.Merge(result[idx:])

				// Cut off all tokens merged into the new token, and append the merged one.
				result = result[:idx]
				result = append(result, mergedToken)
			}
		}
	}

	return result
}
