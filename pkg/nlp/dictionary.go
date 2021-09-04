package nlp

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ReadDictionary reads a Dictionary proto from a file.
// Returns an empty dictionary if path is the empty string.
func ReadDictionary(path string) (*Dictionary, error) {
	dictionary := new(Dictionary)
	if path == "" {
		return dictionary, nil
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	switch ext := filepath.Ext(path); ext {
	case ".pb":
		err = proto.Unmarshal(bytes, dictionary)
		if err != nil {
			return nil, err
		}
	case ".json":
		err = protojson.Unmarshal(bytes, dictionary)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported proto extension %q", ext)
	}

	return dictionary, nil
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
