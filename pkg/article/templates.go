package article

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	TemplateStartPattern = regexp.MustCompile(`\{\{[^#<>[\]|{}]+\|?`)
	TemplateEndPattern   = regexp.MustCompile(`}}`)
)

type TemplateStart string

func (t TemplateStart) Original() string {
	return string(t)
}

func ParseTemplateStart(s string) Token {
	return TemplateStart(s)
}

type TemplateEnd struct{}

func (t TemplateEnd) Original() string {
	return "}}"
}

func ParseTemplateEnd(string) Token {
	return TemplateEnd{}
}

type Template struct {
	Name      string
	Arguments map[string][]Token
}

func (t Template) Original() string {
	switch t.Name {
	case "Blockquote":
		return renderBlockquote(t.Arguments)
	case "IPA-de":
		return renderIPADe(t.Arguments)
	case "IPAc-en":
		return renderIPAcEn(t.Arguments)
	default:
		return ""
	}
}

func renderBlockquote(args map[string][]Token) string {
	var text string
	if textVal, hasText := args["text"]; hasText {
		text = Render(textVal)
	} else if textVal, hasText = args["1"]; hasText {
		text = Render(textVal)
	}

	return "\n" + text + "\n"
}

func renderIPAcEn(args map[string][]Token) string {
	unnamed := 1
	unnamedName := strconv.Itoa(unnamed)

	sb := strings.Builder{}

	for value, exists := args[unnamedName]; exists; {
		for _, t := range value {
			sb.WriteString(t.Original())
		}
		if unnamed == 1 {
			sb.WriteString(": /")
		}

		unnamed++
		unnamedName = strconv.Itoa(unnamed)
		value, exists = args[unnamedName]
	}
	sb.WriteString("/")

	return sb.String()
}

func renderIPADe(args map[string][]Token) string {
	transcription := args["1"]
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

func (t TemplateEnd) Backtrack(tokens []Token) (Token, int) {
	start, startIdx, found := BacktrackUntil[TemplateStart](tokens)
	if !found {
		return nil, startIdx
	}

	name := string(start)
	// Get rid of first two curly braces.
	name = name[2:]
	// Get rid of argument marker if present.
	name = strings.TrimRight(name, "|")

	args := parseArguments(tokens[startIdx+1:])

	return Template{
		Name:      name,
		Arguments: args,
	}, startIdx
}

func parseArguments(tokens []Token) map[string][]Token {
	arguments := make(map[string][]Token)

	var argument []Token
	unnamed := 1
	for idx, token := range tokens {
		literal, isLiteral := token.(LiteralText)
		if !isLiteral {
			argument = append(argument, token)
			continue
		}

		tokenArgs := strings.Split(string(literal), "|")
		for _, tokenArg := range tokenArgs {
			argument = append(argument, LiteralText(tokenArg))
			if len(tokenArgs) > 1 || idx == len(tokens)-1 {
				name, value := parseArgument(argument)

				if name == "" {
					name = strconv.Itoa(unnamed)
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
	wikitext := Render(tokens)
	if !strings.Contains(wikitext, "=") {
		tokens2 := Tokenize(UnparsedText(wikitext))

		return "", tokens2
	}

	splits := strings.Split(wikitext, "=")
	name := splits[0]
	value := splits[1]
	tokens = Tokenize(UnparsedText(value))

	return name, tokens
}
