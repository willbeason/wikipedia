package classify

import (
	"fmt"
	"github.com/willbeason/wikipedia/pkg/documents"
)

func (x *PageClassificationsMap) AddPage(known map[uint32]Classification, pageTitles map[uint32]string, pageId uint32, categories []uint32, pageCategories *documents.PageCategories) {
	x.addPage(known, "", pageTitles, pageId, categories, pageCategories)
}

func (x *PageClassificationsMap) addPage(known map[uint32]Classification, stack string, pageTitles map[uint32]string, pageId uint32, categories []uint32, pageCategories *documents.PageCategories) []Classification {
	stack += fmt.Sprintf(" -> %s", pageTitles[pageId])
	if p, ok := x.Pages[pageId]; ok {
		// We've already gotten the classifications for this page.
		return p.Classifications
	}

	page := &PageClassifications{
		Classifications: make([]Classification, len(Classification_value)),
	}

	// Preemptively assign the page so that we don't get any information from loops of category pages.
	x.Pages[pageId] = page

	for _, categoryId := range categories {
		page.merge(x.addPage(known, stack, pageTitles, categoryId, pageCategories.Pages[categoryId].Categories, pageCategories))
	}

	return page.Classifications
}

func (x *PageClassifications) merge(from []Classification) {
	for i, l := range from {
		x.Classifications[i] |= l
	}
}
