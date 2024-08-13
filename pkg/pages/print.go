package pages

import (
	"fmt"

	"github.com/willbeason/wikipedia/pkg/documents"
)

func Print(page *documents.Page) error {
	fmt.Println(page.Id)
	fmt.Println(page.Title)
	fmt.Println(page.Text)
	return nil
}

func Compare(beforePages <-chan *documents.Page) func(page *documents.Page) error {
	before := make(map[uint32]*documents.Page)

	for page := range beforePages {
		before[page.Id] = page
	}

	return func(page *documents.Page) error {
		gotBefore := before[page.Id]

		fmt.Println(gotBefore.Text)
		fmt.Println(page.Text)

		return nil
	}
}
