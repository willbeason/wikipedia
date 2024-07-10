package protos

import (
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func Read(file string, out proto.Message) error {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("reading %q: %w", file, err)
	}

	switch ext := filepath.Ext(file); ext {
	case ".pb":
		err = proto.Unmarshal(bytes, out)
	case ".json":
		err = protojson.Unmarshal(bytes, out)
	default:
		panic(fmt.Errorf("unsupported proto exension: %q", ext))
	}

	return fmt.Errorf("reading %q: %w", file, err)
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
		if err != nil {
			return fmt.Errorf("writing %q: %w", path, err)
		}
	case ".json":
		bytes, err = protojson.MarshalOptions{Indent: "  "}.Marshal(p)
		if err != nil {
			return fmt.Errorf("writing %q: %w", path, err)
		}
	default:
		return fmt.Errorf("unsupported proto extension %q", ext)
	}

	err = os.WriteFile(path, bytes, os.ModePerm)
	if err != nil {
		return fmt.Errorf("writing %q: %w", path, err)
	}

	return nil
}
