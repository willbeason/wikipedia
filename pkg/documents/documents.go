package documents

import (
	"google.golang.org/protobuf/proto"
)

func ReadPages(pages chan<- *Page) func([]byte) error {
	return func(bytes []byte) error {
		page := &Page{}

		err := proto.Unmarshal(bytes, page)
		if err != nil {
			return err
		}

		pages <- page

		return nil
	}
}
