package nlp

import (
	"strings"

	"github.com/willbeason/wikipedia/pkg/protos"
)

// ReadDictionary reads a Dictionary proto from a file.
// Returns an empty dictionary if path is the empty string.
func ReadDictionary(path string) (*Dictionary, error) {
	dictionary := new(Dictionary)
	if path == "" {
		return dictionary, nil
	}

	dictionary, err := protos.Read[Dictionary](path)
	if err != nil {
		return nil, err
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
