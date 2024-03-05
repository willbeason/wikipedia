package pages

import (
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/protos"
)

// FilterTitles filters a source channel of pages to only include those with the specified Title.
func FilterTitles(titles []string) func(chan<- protos.ID) jobs.Page {
	titlesMap := make(map[string]bool)
	for _, title := range titles {
		titlesMap[title] = true
	}

	return func(ids chan<- protos.ID) jobs.Page {
		return func(page *documents.Page) error {
			if titlesMap[page.Title] {
				ids <- page
			}

			return nil
		}
	}
}
