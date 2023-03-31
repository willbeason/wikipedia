package protos

import (
	"fmt"
	"path/filepath"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"os"
)

func Read(file string, out proto.Message) error {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	switch ext := filepath.Ext(file); ext {
	case ".pb":
		err = proto.Unmarshal(bytes, out)
	case ".json":
		err = protojson.Unmarshal(bytes, out)
	default:
		panic(fmt.Errorf("unsupported proto exension: %q", ext))
	}

	return err
}
