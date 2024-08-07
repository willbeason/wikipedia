package article

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	TemplateStartPattern = regexp.MustCompile(`\{\{[^#<>[\]|{}]+\|?`)
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
	switch t.Name {
	case "IPAc-en":
		return renderIPAcEn(t.Arguments)
	case "IPA-de":
		return renderIPADe(t.Arguments)
	default:
		return ""
	}
}

func renderIPAcEn(args map[string][]Token) string {
	unnamed := 1
	unnamedName := fmt.Sprint(unnamed)

	sb := strings.Builder{}

	for value, exists := args[unnamedName]; exists; {
		for _, t := range value {
			sb.WriteString(t.Render())
		}
		if unnamed == 1 {
			sb.WriteString(": /")
		}

		unnamed++
		unnamedName = fmt.Sprint(unnamed)
		value, exists = args[unnamedName]
	}
	sb.WriteString("/")

	return sb.String()
}

func renderIPADe(args map[string][]Token) string {
	transcription, _ := args["1"]
	display, hasDisplay := args["2"]

	sb := strings.Builder{}
	if hasDisplay {
		renderedDisplay := Render(display)
		switch renderedDisplay {
		case "lang":
			sb.WriteString("German: ")
		default:
			sb.WriteString("German pronunciation: ")
		}
	}

	sb.WriteString("[")
	sb.WriteString(Render(transcription))
	sb.WriteString("]")

	return sb.String()
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

		name := string(*lastStart)
		// Get rid of first two curly braces.
		name = name[2:]
		// Get rid of argument marker if present.
		name = strings.TrimRight(name, "|")

		result = append(result, Template{
			Name:      name,
			Arguments: parseArguments(tokens[lastStartIdx+1 : idx]),
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
	arguments := make(map[string][]Token)

	var argument []Token
	unnamed := 1
	for _, token := range tokens {
		literal, isLiteral := token.(LiteralText)
		if !isLiteral {
			argument = append(argument, token)
			continue
		}

		tokenArgs := strings.Split(string(literal), "|")
		for _, tokenArg := range tokenArgs {
			argument = append(argument, LiteralText(tokenArg))
			if len(tokenArgs) > 1 {
				name, value := parseArgument(argument)

				if name == "" {
					name = fmt.Sprint(unnamed)
					unnamed++
				}

				arguments[name] = value

				argument = nil

			}
		}
	}

	return arguments
}

func parseArgument(tokens []Token) (string, []Token) {
	return "", tokens
}
