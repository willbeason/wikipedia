package article

const (
	RefBeginName = "refbegin"
	RefEndName   = "refend"
)

type References struct {
	Tokens []Token
}

func (t References) Render() string {
	return ""
}

func ParseReferences(tokens []Token) References {
	return References{Tokens: tokens[1 : len(tokens)-1]}
}

func MergeReferences(tokens []Token) ([]Token, bool, error) {
	var result []Token
	appliedRule := false

	lastStartIdx := -1

	for idx, token := range tokens {
		startTemplate, isTemplate := token.(Template)
		if isTemplate && startTemplate.Name == RefBeginName {
			if lastStartIdx != -1 {
				result = append(result, tokens[lastStartIdx:idx]...)
			}

			lastStartIdx = idx
			continue
		} else if lastStartIdx == -1 {
			result = append(result, token)
			continue
		}

		endTemplate, isEnd := token.(Template)
		if !isEnd || endTemplate.Name != RefEndName {
			continue
		}

		appliedRule = true

		result = append(result, ParseReferences(tokens[lastStartIdx:idx+1]))

		lastStartIdx = -1
	}

	if lastStartIdx != -1 {
		result = append(result, tokens[lastStartIdx:]...)
	}

	return result, appliedRule, nil
}
