package pages

import (
	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/jobs"
)

func Filter() jobs.Page {
	return func(page *documents.Page) error {
		return nil
	}
}
