package protos

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

var ErrUnsupportedProtoExtension = errors.New("unsupported proto extension")

// Read reads a protocol buffer stored in file to a protocol buffer of type OUT.
// OUT must be a type whose pointer receiver is a proto.Message.
func Read[OUT any, POUT interface {
	*OUT
	proto.Message
}](file string) (*OUT, error) {
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

func Write(path string, p proto.Message) error {
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return fmt.Errorf("writing %q: %w", path, err)
	}

	var bytes []byte

	switch ext := filepath.Ext(path); ext {
	case ".pb":
		bytes, err = proto.Marshal(p)
	case ".json":
		bytes, err = protojson.MarshalOptions{Indent: "  "}.Marshal(p)
	case ".txt":
		bytes, err = prototext.MarshalOptions{Indent: "  "}.Marshal(p)
	default:
		return fmt.Errorf("%w: %q", ErrUnsupportedProtoExtension, ext)
	}
	if err != nil {
		return fmt.Errorf("marshalling %q: %w", path, err)
	}

	err = os.WriteFile(path, bytes, os.ModePerm)
	if err != nil {
		return fmt.Errorf("writing %q: %w", path, err)
	}

	return nil
}
