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

		mergedToken, idx := endToken.Backtrack(result)
		if idx >= 0 {
			// Cut off all tokens merged into the new token, and append the merged one.
			result = result[:idx]
			result = append(result, mergedToken)
		} else {
			result = append(result, token)
		}
	}

	return result
}
