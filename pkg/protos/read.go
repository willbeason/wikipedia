package protos

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func Read(file string, out proto.Message) error {
	bytes, err := ioutil.ReadFile(file)
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

func Write(path string, p proto.Message) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}

	var bytes []byte

	switch ext := filepath.Ext(path); ext {
	case ".pb":
		bytes, err = proto.Marshal(p)
		if err != nil {
			return err
		}
	case ".json":
		bytes, err = protojson.MarshalOptions{Indent: "  "}.Marshal(p)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported proto extension %q", ext)
	}

	return ioutil.WriteFile(path, bytes, os.ModePerm)
}
