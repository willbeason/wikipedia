package classify

import (
	"github.com/willbeason/wikipedia/pkg/documents"
)

func (x *PageClassificationsMap) AddPage(
	pageTitles map[uint32]string,
	pageID uint32,
	categories []uint32,
	pageCategories *documents.PageCategories,
) {
	x.addPage("", pageTitles, pageID, categories, pageCategories)
}

func (x *PageClassificationsMap) addPage(
	stack string,
	pageTitles map[uint32]string,
	pageID uint32,
	categories []uint32,
	pageCategories *documents.PageCategories,
) []Classification {
	stack += " -> " + pageTitles[pageID]
	if p, ok := x.Pages[pageID]; ok {
		// We've already gotten the classifications for this page.
		return p.Classifications
	}

	page := &PageClassifications{
		Classifications: make([]Classification, len(Classification_value)),
	}

	// Preemptively assign the page so that we don't get any information from loops of category pages.
	x.Pages[pageID] = page

	for _, categoryID := range categories {
		page.merge(x.addPage(stack, pageTitles, categoryID, pageCategories.Pages[categoryID].Categories, pageCategories))
	}

	return page.Classifications
}

func (x *PageClassifications) merge(from []Classification) {
	for i, l := range from {
		x.Classifications[i] |= l
	}
}
