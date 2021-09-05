package documents

import (
	"fmt"
	"github.com/willbeason/wikipedia/pkg/documents/tagtree"
	"regexp"
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

	const title = "Category:Evangelicalism in Austria"

	idx := 0
	for _, line := range categoryLines {
		//if strings.Contains(line, "MONTHNUMBER") {
		//	panic(fmt.Sprintf("(%q, %q)\n\n", page.Title, line))
		//}

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
			categoryId, ok = c.TitleIndex.Titles[strings.ReplaceAll(categoryTitle, "-", " ")]
		}

		if !ok {
			//if page.Title == title {
				fmt.Printf("(%q, %q)\n%s\n\n", page.Title, line, categoryTitle)
			//}


			// People may misspell categories.
			Missed++

			continue
		}

		result.Categories[idx] = categoryId
		idx++
	}

	//if page.Title == title {
	//	panic("DONE")
	//}

	result.Categories = result.Categories[:idx]

	return result
}

var spaces = regexp.MustCompile(`\s+`)

func DisambiguateTags(page, category string) string {
	category = strings.ReplaceAll(category, "_", " ")
	category = spaces.ReplaceAllString(category, " ")
	category = strings.ReplaceAll(category, "&ndash;", "–")

	t := tagtree.Parse(category)

	return t.String(page)
}
