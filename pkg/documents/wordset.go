package documents

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type WordSets struct {
	InFile    string `json:"in_file,omitempty"`
	Documents []WordSet
}

type WordSet struct {
	// ID is the article ID
	ID uint32
	// Words is the sorted list of top words in the document.
	Words []uint32
}

func ReadWordSet(line string) (*WordSet, error) {
	parts := strings.Split(line, ":")
	if len(parts) != 2 {
		return nil, errors.New("invalid line")
	}

	id, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return nil, err
	}

	wordStrings := strings.Split(parts[1], ",")
	words := make([]uint32, len(wordStrings))

	for i, w := range wordStrings {
		n, err := strconv.ParseUint(w, 10, 32)
		if err != nil {
			return nil, err
		}
		words[i] = uint32(n)
	}

	return &WordSet{ID: uint32(id), Words: words}, nil
}

func (ws WordSet) String() string {
	words := make([]string, len(ws.Words))
	for i, n := range ws.Words {
		words[i] = fmt.Sprint(n)
	}
	wordsPart := strings.Join(words, ",")

	return fmt.Sprintf("%d:%s", ws.ID, wordsPart)
}
