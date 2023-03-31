package nlp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	protobuf = ".pb"
	json     = ".json"
)

// ReadDictionary reads a Dictionary proto from a file.
// Returns an empty dictionary if path is the empty string.
func ReadDictionary(path string) (*Dictionary, error) {
	dictionary := new(Dictionary)
	if path == "" {
		return dictionary, nil
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	switch ext := filepath.Ext(path); ext {
	case protobuf:
		err = proto.Unmarshal(bytes, dictionary)
		if err != nil {
			return nil, err
		}
	case json:
		err = protojson.Unmarshal(bytes, dictionary)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported proto extension %q", ext)
	}

	return dictionary, nil
}

func WriteDictionary(path string, dictionary *Dictionary) error {
	var bytes []byte
	var err error

	switch ext := filepath.Ext(path); ext {
	case protobuf:
		bytes, err = proto.Marshal(dictionary)
		if err != nil {
			return err
		}
	case json:
		bytes, err = protojson.MarshalOptions{Indent: "  "}.Marshal(dictionary)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported proto extension %q", ext)
	}

	return os.WriteFile(path, bytes, os.ModePerm)
}

func ToNgramDictionary(dictionary *Dictionary) map[string]bool {
	result := make(map[string]bool, len(dictionary.Words))

	for _, word := range dictionary.Words {
		words := strings.Split(word, " ")
		for i := 1; i <= len(words); i++ {
			ngram := strings.Join(words[:i], " ")
			result[ngram] = true
		}
	}

	return result
}
