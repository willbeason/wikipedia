package articles

import (
	"errors"
)

var ErrTokenize = errors.New("tokenizing article")

func Tokenize(text UnparsedText) ([]Token, error) {
	tokens := []Token{text}

	oneTimeRules := []RuleFn{
		NowikiSectionTokens,
		PatternTokenRule(NowikiAutoClosePattern, func(string) Token {
			return NowikiAutoClose{}
		}),
		PatternTokenRule(TemplateStartPattern, func(s string) Token {
			return TemplateStart(s)
		}),
		PatternTokenRule(TemplateEndPattern, func(string) Token {
			return TemplateEnd{}
		}),
		ToLiterals,
	}

	var err error
	for _, rule := range oneTimeRules {
		tokens, _, err = rule(tokens)
		if err != nil {
			return nil, err
		}
	}

	repeatedRules := []RuleFn{
		MergeTemplateTokens,
	}

	for _, rule := range repeatedRules {
		// apply the rule at least once.
		applied := true
		for applied {
			tokens, applied, err = rule(tokens)
			if err != nil {
				return nil, err
			}
		}
	}

	return tokens, nil
}
