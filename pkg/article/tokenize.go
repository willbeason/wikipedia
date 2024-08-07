package article

import (
	"errors"
)

var ErrTokenize = errors.New("tokenizing article")

func Tokenize(text UnparsedText) ([]Token, error) {
	tokens := []Token{text}

	oneTimeRules := []RuleFn{
		NowikiSectionTokens,
		PatternTokenRule(NowikiAutoClosePattern, ParseNowikiAutoClose),
		PatternTokenRule(TemplateStartPattern, ParseTemplateStart),
		PatternTokenRule(TemplateEndPattern, ParseTemplateEnd),
		PatternTokenRule(RefAutoClosePattern, ParseRefAutoClose),
		PatternTokenRule(RefStartPattern, ParseRefStart),
		PatternTokenRule(RefEndPattern, ParseRefEnd),
		PatternTokenRule(LinkStartPattern, ParseLinkStart),
		PatternTokenRule(LinkEndPattern, ParseLinkEnd),
		PatternTokenRule(BlockquoteStartPattern, ParseBlockquoteStart),
		PatternTokenRule(BlockquoteEndPattern, ParseBlockquoteEnd),
		PatternTokenRule(EmphasisStartPattern, ParseEmphasisStart),
		PatternTokenRule(EmphasisEndPattern, ParseEmphasisEnd),
		PatternTokenRule(MathStartPattern, ParseMathStart),
		PatternTokenRule(MathEndPattern, ParseMathEnd),
		PatternTokenRule(SubscriptStartPattern, ParseSubscriptStart),
		PatternTokenRule(SubscriptEndPattern, ParseSubscriptEnd),
		PatternTokenRule(ExternalLinkPattern, ParseExternalLink),
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
		MergeTokenRule[BlockquoteStart, BlockquoteEnd](ParseBlockquote),
		MergeTokenRule[EmphasisStart, EmphasisEnd](ParseEmphasis),
		MergeTokenRule[MathStart, MathEnd](ParseMath),
		MergeTokenRule[SubscriptStart, SubscriptEnd](ParseSubscript),
		MergeReferences,
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
