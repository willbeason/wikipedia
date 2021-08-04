package nlp

import (
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

type Dictionary map[string]bool

func ReadDictionary(path string) (Dictionary, error) {
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

func WriteDictionary(path string, dictionary Dictionary) error {
	words := make([]string, len(dictionary))
	idx := 0

	for w := range dictionary {
		words[idx] = w
		idx++
	}

	sort.Strings(words)

	out, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func() {
		// For deferred closing, we don't care about the error as we've already
		// encountered one.
		_ = out.Close()
	}()

	for _, word := range words {
		_, err = out.WriteString(word)
		if err != nil {
			return err
		}

		_, err = out.WriteString("\n")
		if err != nil {
			return err
		}
	}

	err = out.Sync()
	if err != nil {
		return err
	}

	return out.Close()
}
