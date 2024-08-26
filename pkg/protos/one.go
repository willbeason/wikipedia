package protos

import (
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

// ReadOne reads a protocol buffer stored in file to a protocol buffer of type OUT.
func ReadOne[OUT any, POUT Proto[OUT]](file string) (*OUT, error) {
	var out POUT = new(OUT)

	bytes, err := os.ReadFile(file)
	if err != nil {
		return out, fmt.Errorf("reading %q: %w", file, err)
	}

	switch ext := filepath.Ext(file); ext {
	case ".pb":
		err = proto.Unmarshal(bytes, out)
	case ".json":
		err = protojson.Unmarshal(bytes, out)
	case ".txt":
		err = prototext.Unmarshal(bytes, out)
	default:
		return out, fmt.Errorf("%w: %q", ErrUnsupportedProtoExtension, ext)
	}

	if err != nil {
		return out, fmt.Errorf("unmarshalling %q: %w", file, err)
	}

	return out, nil
}
