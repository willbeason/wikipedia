package documents

import (
	"context"
	"github.com/willbeason/wikipedia/pkg/jobs"
)

func NewTitleMap() map[string]uint32 {
	return make(map[string]uint32)
}

func MakeTitleMap(title *ArticleIdTitle, titleMap map[string]uint32) error {
	titleMap[title.Title] = title.Id
	return nil
}

func MakeTitleMapFn(titles <-chan *ArticleIdTitle, titleMap chan<- map[string]uint32) jobs.Job {
	return func(ctx context.Context, _ chan<- error) {
		result := make(map[string]uint32)
		defer func() {
			titleMap <- result
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case title, ok := <-titles:
				if !ok {
					return
				}

				result[title.Title] = title.Id
			}
		}
	}
}
