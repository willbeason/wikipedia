package documents

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

func ReadPages[T any, PT interface {
	*T
	proto.Message
}](pages chan<- PT) func([]byte) error {
	return func(bytes []byte) error {
		var page PT = new(T)

		err := proto.Unmarshal(bytes, page)
		if err != nil {
			return fmt.Errorf("unmarshalling page: %w", err)
		}

		pages <- page

		return nil
	}
}
