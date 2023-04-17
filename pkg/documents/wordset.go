package documents

import (
	"errors"
	"fmt"
	"os"
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

func (ws *WordSet) ToSet() map[uint32]bool {
	result := make(map[uint32]bool, len(ws.Words))

	for _, w := range ws.Words {
		result[w] = true
	}

	return result
}

func (ws *WordSet) ToBits(length int) []bool {
	result := make([]bool, length)

	for _, w := range ws.Words {
		result[w] = true
	}

	return result
}

func ReadWordSets(path string) ([]WordSet, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(bytes), "\n")
	result := make([]WordSet, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		ws, err2 := ReadWordSet(line)
		if err2 != nil {
			return nil, err2
		}

		result = append(result, ws)
	}

	return result, nil
}

func ReadWordSet(line string) (WordSet, error) {
	parts := strings.Split(line, ":")
	if len(parts) != 2 {
		return WordSet{}, errors.New("invalid line for word set")
	}

	id, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return WordSet{}, err
	}

	wordStrings := strings.Split(parts[1], ",")
	words := make([]uint32, len(wordStrings))

	for i, w := range wordStrings {
		n, err := strconv.ParseUint(w, 10, 32)
		if err != nil {
			return WordSet{}, err
		}
		words[i] = uint32(n)
	}

	return WordSet{ID: uint32(id), Words: words}, nil
}

func (ws WordSet) String() string {
	words := make([]string, len(ws.Words))
	for i, n := range ws.Words {
		words[i] = fmt.Sprint(n)
	}
	wordsPart := strings.Join(words, ",")

	return fmt.Sprintf("%d:%s", ws.ID, wordsPart)
}
