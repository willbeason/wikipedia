package nlp

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Gender string

const (
	Male      Gender = "male"
	Female    Gender = "female"
	Nonbinary Gender = "nonbinary"
	Multiple  Gender = "multiple"
	Unknown   Gender = "unknown"
)

type DocumentGender struct {
	ID     uint32
	Gender Gender
}

func (dg DocumentGender) String() string {
	return fmt.Sprintf("%d:%s", dg.ID, dg.Gender)
}

func ReadDocumentGenders(path string) ([]DocumentGender, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(bytes), "\n")
	result := make([]DocumentGender, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		ws, err2 := ReadDocumentGender(line)
		if err2 != nil {
			return nil, err2
		}

		result = append(result, ws)
	}

	return result, nil
}

func ReadDocumentGender(line string) (DocumentGender, error) {
	parts := strings.Split(line, ":")
	if len(parts) != 2 {
		return DocumentGender{}, errors.New("invalid line for document gender")
	}

	id, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return DocumentGender{}, err
	}

	g := Gender(parts[1])
	switch g {
	case Male, Female, Nonbinary, Multiple, Unknown:
	default:
		return DocumentGender{}, errors.New("unknown gender label")
	}

	return DocumentGender{ID: uint32(id), Gender: g}, nil
}

var (
	categoryRegex  = regexp.MustCompile(`\\[\\[Category:.+]]`)
	womenRegex     = regexp.MustCompile(`\\b(women|female)\\b`)
	menRegex       = regexp.MustCompile(`\\b(men|male)\\b`)
	nonbinaryRegex = regexp.MustCompile(`\b(nonbinary)\b`)

	femalePronouns    = regexp.MustCompile(`\\b(she|hers|her|herself)\\b`)
	malePronouns      = regexp.MustCompile(`\\b(he|his|him|himself)\\b`)
	nonbinaryPronouns = regexp.MustCompile(`\\b(they|their|theirs|them|themself)\\b`)
)

func DetermineGender(text string) Gender {
	categories := categoryRegex.FindAllString(text, -1)

	foundMale := false
	foundFemale := false
	foundNonbinary := false

	for _, category := range categories {
		category = strings.ToLower(category)
		if womenRegex.MatchString(category) {
			foundFemale = true
		} else if menRegex.MatchString(category) {
			foundMale = true
		} else if nonbinaryRegex.MatchString(category) {
			foundNonbinary = true
		}
	}

	cleanedText := strings.ToLower(CleanArticle(text))
	//firstFemale := femalePronouns.FindAllStringIndex(cleanedText, 1)
	//firstFemaleIdx := math.MaxInt
	//if len(firstFemale) > 0 {
	//	firstFemaleIdx = firstFemale[0][0]
	//}
	//
	//firstMale := malePronouns.FindAllStringIndex(cleanedText, 1)
	//firstMaleIdx := math.MaxInt
	//if len(firstMale) > 0 {
	//	firstMaleIdx = firstMale[0][0]
	//}

	//firstNonbinary := nonbinaryPronouns.FindAllStringIndex(cleanedText, 1)
	//firstNonbinaryIdx := math.MaxInt
	//if len(firstNonbinary) > 0 {
	//	firstNonbinaryIdx = firstNonbinary[0][0]
	//}

	//switch {
	//case firstFemaleIdx < firstMaleIdx:
	//	foundFemale = true
	//case firstMaleIdx < firstFemaleIdx:
	//	foundMale = true
	//}

	femaleUsages := len(femalePronouns.FindAllString(cleanedText, -1))
	maleUsages := len(malePronouns.FindAllString(cleanedText, -1))
	nonbinaryUsages := len(nonbinaryPronouns.FindAllString(cleanedText, -1))

	switch {
	case femaleUsages > maleUsages && femaleUsages > nonbinaryUsages:
		foundFemale = true
	case maleUsages > femaleUsages && maleUsages > nonbinaryUsages:
		foundMale = true
	case nonbinaryUsages > femaleUsages && nonbinaryUsages > maleUsages:
		foundNonbinary = true
	}

	switch {
	case foundMale && foundFemale || foundMale && foundNonbinary || foundFemale && foundNonbinary:
		return Multiple
	case foundMale:
		return Male
	case foundFemale:
		return Female
	case foundNonbinary:
		return Nonbinary
	}

	// No signals.
	return Unknown
}
