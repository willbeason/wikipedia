package ordinality

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/willbeason/extract-wikipedia/pkg/nlp"
	"google.golang.org/protobuf/proto"
)

// WordOrdinality is a map from a word (or n-gram) to its ordinal position in a
// dictionary. The positions are 1-based so 0 (the null value) can be used to
// represent words not in the dictionary.
type WordOrdinality map[string]uint32

func NewWordOrdinality(dictionary *nlp.Dictionary) WordOrdinality {
	if dictionary == nil {
		return nil
	}
	result := make(WordOrdinality, len(dictionary.Words))

	for i, word := range dictionary.Words {
		result[word] = uint32(i) + 1
	}

	return result
}

func WriteWordBags(path string, wordBags *DocumentWordBag) error {
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return err
	}

	bytes, err := proto.Marshal(wordBags)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, bytes, os.ModePerm)
}

// ReadWordBags reads a Dictionary proto from a file.
// Returns an empty dictionary if path is the empty string.
func ReadWordBags(path string) (*DocumentWordBag, error) {
	wordBags := new(DocumentWordBag)
	if path == "" {
		return wordBags, nil
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = proto.Unmarshal(bytes, wordBags)
	if err != nil {
		return nil, err
	}

	return wordBags, nil
}
