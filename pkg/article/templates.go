package articles

import "regexp"

var (
	TemplateStartPattern = regexp.MustCompile(`\{\{[^#<>[\]|{}]+`)
	TemplateEndPattern   = regexp.MustCompile(`}}`)
)

type TemplateStart string

func (t TemplateStart) Render() string {
	return string(t)
}

type TemplateEnd struct{}

func (t TemplateEnd) Render() string {
	return "}}"
}

type Template struct {
	Name      string
	Arguments map[string][]Token
}

func (t Template) Render() string {
	return ""
}

func MergeTemplateTokens(tokens []Token) ([]Token, bool, error) {
	var result []Token
	appliedRule := false

	lastStartIdx := -1
	var lastStart *TemplateStart

	for idx, token := range tokens {
		start, isStart := token.(TemplateStart)
		if isStart {
			if lastStartIdx != -1 {
				result = append(result, tokens[lastStartIdx:idx]...)
			}

			lastStart = &start
			lastStartIdx = idx
			continue
		} else if lastStart == nil {
			// No template to close even if this is an end token.
			result = append(result, token)
			continue
		}

		_, isEnd := token.(TemplateEnd)
		if !isEnd {
			continue
		}

		// We have a start token and the next template token is an end.
		appliedRule = true
		result = append(result, Template{
			Name:      string(*lastStart)[2:],
			Arguments: parseArguments(tokens[lastStartIdx:idx]),
		})

		lastStart = nil
		lastStartIdx = -1
	}

	// Fill in remaining tokens.
	if lastStartIdx != -1 {
		result = append(result, tokens[lastStartIdx:]...)
	}

	return result, appliedRule, nil
}

func parseArguments(tokens []Token) map[string][]Token {
	return nil
}
