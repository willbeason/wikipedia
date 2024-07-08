package documents

import (
	"regexp"
	"strings"

	"github.com/willbeason/wikipedia/pkg/documents/tagtree"
)

type Categorizer struct {
	TitleIndex *TitleIndex
}

var categorySplit = regexp.MustCompile(`]]\s*\[\[`)

var Missed = 0

func (c *Categorizer) Categorize(page *Page) *Categories {
	const (
		categoryPrefix     = "[[Category:"
		categoryTrimPrefix = "[["
		categorySuffix     = "]]"
	)

	text := categorySplit.ReplaceAllString(page.Text, "]]\n[[")
	lines := strings.Split(text, "\n")

	//nolint: prealloc // Cost of estimating is higher than potential gain.
	// Consider setting at median size to reduce reallocations.
	var categoryLines []string

	for _, line := range lines {
		if !strings.HasPrefix(line, categoryPrefix) {
			continue
		}
		line = strings.TrimPrefix(line, categoryTrimPrefix)

		if !strings.HasSuffix(line, categorySuffix) {
			continue
		}

		line = strings.TrimSuffix(line, categorySuffix)
		categoryLines = append(categoryLines, line)
	}

	result := &Categories{Categories: make([]uint32, len(categoryLines))}

	idx := 0
	for _, line := range categoryLines {
		categoryTitle := line

		categoryTitle = DisambiguateTags(page.Title, categoryTitle)

		bar := strings.Index(categoryTitle, "|")
		if bar != -1 {
			categoryTitle = categoryTitle[:bar]
		}

		categoryTitle = strings.ToLower(categoryTitle)
		categoryTitle = strings.TrimSpace(categoryTitle)
		if categoryTitle == "category:" || categoryTitle == "category:{category}" {
			continue
		}

		if strings.HasPrefix(categoryTitle, "category: ") {
			categoryTitle = "category:" + categoryTitle[10:]
		}

		if strings.Contains(categoryTitle, "different number of open and close braces") {
			continue
		}

		categoryID, ok := c.TitleIndex.Titles[categoryTitle]

		if !ok {
			categoryID, ok = c.TitleIndex.Titles[strings.ReplaceAll(categoryTitle, "-", " ")]
		}

		if !ok {
			// People may misspell categories.
			Missed++

			continue
		}

		result.Categories[idx] = categoryID
		idx++
	}

	result.Categories = result.Categories[:idx]

	return result
}

var (
	spaces  = regexp.MustCompile(`\s+`)
	comment = regexp.MustCompile(`<!--[^>]+-->`)

	ignoredCharacters = regexp.MustCompile("[\u200e\u202a\u202c]")
)

func DisambiguateTags(page, category string) string {
	category = strings.ReplaceAll(category, "_", " ")
	category = spaces.ReplaceAllString(category, " ")
	category = strings.ReplaceAll(category, "&ndash;", "–")
	category = ignoredCharacters.ReplaceAllString(category, "")
	category = comment.ReplaceAllString(category, "")
	category = strings.ReplaceAll(category, "{{!}}", "|")

	t := tagtree.Parse(category)

	return t.String(page)
}
