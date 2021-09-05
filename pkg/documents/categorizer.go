package documents

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Categorizer struct {
	TitleIndex *TitleIndex
}

// Categorize returns the Categories a Page identifies itself as having.
//
// Sample Category lines:
// - [[Category:Main topic classifications]]

var Missed = 0

func (c *Categorizer) Categorize(page *Page) *Categories {
	const (
		categoryPrefix     = "[[Category:"
		categoryTrimPrefix = "[["
		categorySuffix     = "]]"
	)

	text := strings.ReplaceAll(page.Text, "]][[", "]]\n[[")
	lines := strings.Split(text, "\n")

	var categoryLines []string

	for _, line := range lines {
		if !strings.HasPrefix(line, categoryPrefix) {
			continue
		}
		line = strings.TrimPrefix(line, categoryTrimPrefix)

		if strings.HasPrefix(line, "Category: ") {
			line = strings.Replace(line, "Category: ", "Category:", 1)
		}

		if !strings.HasSuffix(line, categorySuffix) {
			continue
		}

		line = strings.TrimSuffix(line, categorySuffix)
		categoryLines = append(categoryLines, line)
	}

	result := &Categories{Categories: make([]uint32, len(categoryLines))}

	idx := 0
	for _, line := range categoryLines {
		//fmt.Println(line)
		categoryTitle := line

		categoryTitle = DisambiguateTags(page.Title, categoryTitle)

		bar := strings.Index(categoryTitle, "|")
		if bar != -1 {
			categoryTitle = categoryTitle[:bar]
		}

		categoryTitle = strings.ToLower(categoryTitle)
		categoryTitle = strings.TrimSpace(categoryTitle)
		if categoryTitle == "category:" {
			continue
		}

		categoryId, ok := c.TitleIndex.Titles[categoryTitle]
		if !ok {
			// People may misspell categories.
			Missed++
			fmt.Printf("(%q, %q)\n%s\n\n", page.Title, line, categoryTitle)
		}

		result.Categories[idx] = categoryId
		idx++
	}

	if page.Id == 58848427 {
		fmt.Println(result.Categories)
	}

	result.Categories = result.Categories[:idx]

	return result
}

var (
	yearPattern      = regexp.MustCompile(`\d{3,4}`)
	monthPattern     = regexp.MustCompile(`(?i)(january|february|march|april|may|june|july|august|september|october|november|december)`)
	decadePattern    = regexp.MustCompile(`\d{2,3}0s`)
	countryPattern   = regexp.MustCompile(`\bin( \w+)+$`)
	yearRangePattern = regexp.MustCompile(`\d{4}–\d{2}`)
	centuryPattern   = regexp.MustCompile(`(\d{1,2}(st|nd|rd|th))[ -]century`)

	titleYearTag      = regexp.MustCompile(`(?i){{title year}}`)
	titleMonthTag     = regexp.MustCompile(`(?i){{title monthname}}`)
	titleDecadeTag    = regexp.MustCompile(`(?i){{title decade}}s?`)
	titleCountryTag   = regexp.MustCompile(`(?i)in {{title country}}`)
	titleYearRangeTag = regexp.MustCompile(`(?i){{title year range}}`)

	inferMonthTag          = regexp.MustCompile(`(?i){{MONTH\|\w+}}`)
	inferDecadeTag         = regexp.MustCompile(`(?i){{DECADE\|(\d+)}}`)
	inferCenturyTag        = regexp.MustCompile(`(?i){{(century from year|century name from title year)(\|\d{3,4})?(\|dash)?}}`)
	inferCenturyTag2       = regexp.MustCompile(`(?i){{(century from decade|century name from title decade)(\|\d{3,4})?(\|dash)?}}`)
	inferOrdinalCenturyTag = regexp.MustCompile(`(?i){{Ordinal\|{{title century}}}}`)

	spaces = regexp.MustCompile(`\s+`)
)

func DisambiguateTags(page, category string) string {
	category = strings.ReplaceAll(category, "_", " ")
	category = spaces.ReplaceAllString(category, " ")

	category = disambiguateTag(titleYearTag, yearPattern, page, category)
	category = disambiguateTag(titleMonthTag, monthPattern, page, category)
	category = disambiguateTag(inferMonthTag, monthPattern, page, category)
	category = disambiguateTag(titleDecadeTag, decadePattern, page, category)
	category = disambiguateTag(titleYearRangeTag, yearRangePattern, page, category)

	category = disambiguateCountry(page, category)

	category = disambiguateDecade(category)
	category = disambiguateCenturyFromYear(page, category)
	category = disambiguateCenturyFromDecade(page, category)
	category = disambiguateOrdinalCentury(page, category)

	return category
}

func disambiguateCountry(page, category string) string {
	if !titleCountryTag.MatchString(category) {
		return category
	}

	countries := countryPattern.FindAllString(page, -1)
	if len(countries) != 1 {
		return category
	}

	country := countries[0]

	country = strings.TrimSuffix(country, " by month")

	return titleCountryTag.ReplaceAllString(category, country)
}

func disambiguateTag(tag, pattern *regexp.Regexp, page, category string) string {
	if !tag.MatchString(category) {
		return category
	}

	patternMatches := pattern.FindAllString(page, -1)
	if len(patternMatches) != 1 {
		return category
	}

	year := patternMatches[0]

	return tag.ReplaceAllString(category, year)
}

func disambiguateDecade(category string) string {
	years := inferDecadeTag.FindAllStringSubmatch(category, -1)
	if len(years) != 1 {
		return category
	}

	year := years[0][1]
	decade := year[:len(year)-1] + "0s"
	return inferDecadeTag.ReplaceAllString(category, decade)
}

func disambiguateCenturyFromYear(title, category string) string {
	if !inferCenturyTag.MatchString(category) {
		return category
	}

	years := yearPattern.FindAllString(title, -1)
	if len(years) != 1 {
		panic(category)
		return category
	}

	year := years[0]
	centuryStr := year[:len(year)-2]
	century, err := strconv.ParseUint(centuryStr, 10, 8)
	if err != nil {
		return category
	}

	century++
	suffix := "th"
	switch century % 10 {
	case 1:
		if century%100 != 11 {
			suffix = "st"
		}
	case 2:
		if century%100 != 12 {
			suffix = "nd"
		}
	case 3:
		if century%100 != 13 {
			suffix = "rd"
		}
	}

	centuryReplace := fmt.Sprintf("%d%s century", century, suffix)
	if strings.Contains(category, "|dash") {
		centuryReplace = fmt.Sprintf("%d%s-century", century, suffix)
	}

	return inferCenturyTag.ReplaceAllString(category, centuryReplace)
}

func disambiguateCenturyFromDecade(title, category string) string {
	if !inferCenturyTag2.MatchString(category) {
		return category
	}

	decades := decadePattern.FindAllString(title, -1)
	if len(decades) != 1 {
		return category
	}

	decade := decades[0]
	centuryStr := decade[:len(decade)-3]
	century, err := strconv.ParseUint(centuryStr, 10, 8)
	if err != nil {
		return category
	}

	var centuryReplace string
	if strings.Contains(category, "|dash") {
		centuryReplace = fmt.Sprintf("%s-century", toOrdinal(century+1))
	} else {
		centuryReplace = fmt.Sprintf("%s century", toOrdinal(century+1))
	}

	return inferCenturyTag2.ReplaceAllString(category, centuryReplace)
}

func disambiguateOrdinalCentury(title, category string) string {
	if !inferOrdinalCenturyTag.MatchString(category) {
		return category
	}

	centuries := centuryPattern.FindAllStringSubmatch(title, -1)
	if len(centuries) != 1 {
		return category
	}

	return inferOrdinalCenturyTag.ReplaceAllString(category, centuries[0][1])
}

func toOrdinal(i uint64) string {
	suffix := "th"
	switch i % 10 {
	case 1:
		if i%100 != 11 {
			suffix = "st"
		}
	case 2:
		if i%100 != 12 {
			suffix = "nd"
		}
	case 3:
		if i%100 != 13 {
			suffix = "rd"
		}
	}

	return fmt.Sprintf("%d%s", i, suffix)
}
