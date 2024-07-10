package nlp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ReadDictionary reads a Dictionary proto from a file.
// Returns an empty dictionary if path is the empty string.
//
//goland:noinspection GoUnusedExportedFunction
func ReadDictionary(path string) (*Dictionary, error) {
	dictionary := new(Dictionary)
	if path == "" {
		return dictionary, nil
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %q: %w", path, err)
	}

	switch ext := filepath.Ext(path); ext {
	case ".pb":
		err = proto.Unmarshal(bytes, dictionary)
		if err != nil {
			return nil, fmt.Errorf("unmarshalling %q: %w", path, err)
		}
	case ".json":
		err = protojson.Unmarshal(bytes, dictionary)
		if err != nil {
			return nil, fmt.Errorf("unmarshalling %q: %w", path, err)
		}
	default:
		return nil, fmt.Errorf("unsupported proto extension %q", ext)
	}

	return dictionary, nil
}

// ToNgramDictionary constructs a dictionary of all n-grams including prefixes.
func ToNgramDictionary(dictionary *Dictionary) map[string]bool {
	result := make(map[string]bool, len(dictionary.Words))

	for _, word := range dictionary.Words {
		words := strings.Split(word, " ")
		// Include all prefixes of n-grams. Makes no changes to 1-grams.
		for i := 1; i <= len(words); i++ {
			ngram := strings.Join(words[:i], " ")
			result[ngram] = true
		}
	}

	return result
}
