package nlp

import (
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

func ReadDictionary(path string) (map[string]bool, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(bytes), "\n")

	dictionary := make(map[string]bool, len(lines))
	for _, word := range lines {
		if word == "" {
			continue
		}
		dictionary[word] = true
	}

	return dictionary, nil
}

func WriteDictionary(path string, dictionary map[string]bool) error {
	words := make([]string, len(dictionary))

	idx := 0
	for w := range dictionary {
		words[idx] = w
		idx++
	}

	sort.Strings(words)

	return ioutil.WriteFile(path, []byte(strings.Join(words, "\n")), os.ModePerm)
}
