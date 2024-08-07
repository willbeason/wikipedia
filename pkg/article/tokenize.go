package article

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
		PatternTokenRule(RefAutoclosePattern, func(s string) Token {
			return RefAutoclose(s)
		}),
		PatternTokenRule(RefStartPattern, func(s string) Token {
			return RefStart(s)
		}),
		PatternTokenRule(RefEndPattern, func(string) Token {
			return RefEnd{}
		}),
		PatternTokenRule(LinkStartPattern, func(string) Token {
			return LinkStart{}
		}),
		PatternTokenRule(LinkEndPattern, func(string) Token {
			return LinkEnd{}
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
		MergeTokenRule[RefStart, RefEnd](ParseRef),
		MergeTokenRule[LinkStart, LinkEnd](ParseLink),
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
