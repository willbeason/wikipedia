package article

import (
	"errors"
)

var ErrTokenize = errors.New("tokenizing article")

func Tokenize(text UnparsedText) []Token {
	tokens := []Token{text}

	nowikiRules := []RuleFn{
		ExactTokenRule(NowikiSectionStartPattern, func() Token {
			return NowikiStart{}
		}),
		ExactTokenRule(NowikiSectionEndPattern, func() Token {
			return NowikiEnd{}
		}),
		MergeNowikiTokens,
	}

	for _, rule := range nowikiRules {
		tokens = rule(tokens)
	}

	oneTimeRules := []RuleFn{
		PatternTokenRule(NowikiAutoClosePattern, ParseNowikiAutoClose),
		PatternTokenRule(HeaderStartPattern, ParseHeaderStart),
		PatternTokenRule(HeaderEndPattern, ParseHeaderEnd),
		PatternTokenRule(TemplateStartPattern, ParseTemplateStart),
		PatternTokenRule(TemplateEndPattern, ParseTemplateEnd),
		PatternTokenRule(RefAutoClosePattern, ParseRefAutoClose),
		PatternTokenRule(RefStartPattern, ParseRefStart),
		PatternTokenRule(RefEndPattern, ParseRefEnd),
		PatternTokenRule(LinkStartPattern, ParseLinkStart),
		PatternTokenRule(LinkEndPattern, ParseLinkEnd),
		PatternTokenRule(ExternalLinkPattern, ParseExternalLink),
		PatternTokenRule(CommentPattern, ParseComment),
		ExactTokenRule(TableStartPattern, func() Token {
			return TableStart(TableStartPattern)
		}),
		ExactTokenRule(TableEndPattern, func() Token {
			return TableEnd(TableEndPattern)
		}),
		PatternTokenRule(HTMLOpenTagPattern, ParseHTMLOpenTag),
		PatternTokenRule(HTMLCloseTagPattern, ParseHTMLCloseTag),
		ToLiterals,
	}

	for _, rule := range oneTimeRules {
		tokens = rule(tokens)
	}

	tokens = MergeTokens(tokens)

	repeatedRules := []RuleFn{
		MergeReferences,
		MergeSections,
	}

	for _, rule := range repeatedRules {
		tokens = rule(tokens)
	}

	return tokens
}
