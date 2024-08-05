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
