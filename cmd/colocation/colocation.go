package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/walker"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

var stopWords = map[string]bool {
	"noinclude": true,
	"afd": true,
	"wikiproject": true,
	"diff": true,
	"xfd": true,
	"boilerplate": true,
	"overlap": true,
	"notability": true,
	"em": true,
	"php": true,
	"blacklist": true,
	"whitelist": true,
	"td": true,
	"edit": true,
	"disambiguation": true,
	"username": true,
	"admin": true,
	"metadata": true,
	"bot": true,
	"edits": true,
	"tr": true,
	"templates": true,
	"uploaded": true,
	"utc": true,
	"whatlinkshere": true,
	"gng": true,
	"delete": true,
	"overlaps": true,
	"tt": true,
	"deletion": true,
	"coi": true,
	"stub": true,
	"revert": true,
	"nominations": true,
	"afc": true,
	"sorting": true,
	"speedy": true,
	"redirect": true,
	"user": true,
	"https": true,
	"http": true,
	"margin": true,
	"nom": true,
	"mediawiki": true,
	"lx": true,
	"posted": true,
	"manually": true,
	"middot": true,
	"vandalism": true,
	"rfc": true,
	"nominate": true,
	"logs": true,
	"unsigned": true,
	"pov": true,
	"assessed": true,
	"upload": true,
	"htm": true,
	"gt": true,
	"navbox": true,
	"frac": true,
	"uefa": true,
	"notify": true,
	"contribution": true,
	"padding": true,
}

var cmd = cobra.Command{
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		wordSetsDir := args[0]
		dictionaryFile := args[1]

		work := make(chan string)
		errs := make(chan error)

		go func() {
			err := filepath.WalkDir(wordSetsDir, walker.Files(work))
			if err != nil {
				errs <- err
			}
			close(work)
		}()

		workWg := sync.WaitGroup{}

		seenSets := make(chan *documents.WordSets, 0)

		for i := 0; i < 8; i++ {
			workWg.Add(1)
			go func() {
				for item := range work {
					newWordSets, err := readWordSets(item)
					if err != nil {
						errs <- fmt.Errorf("%s: %w", item, err)
						continue
					}
					seenSets <- newWordSets
				}
				workWg.Done()
			}()
		}

		errsWg := sync.WaitGroup{}
		errsWg.Add(1)
		go func() {
			for err := range errs {
				fmt.Println(err)
			}
			errsWg.Done()
		}()

		allSeenParts := make([][]int, 8)
		seenWg := sync.WaitGroup{}

		for i := 0; i < 8; i++ {
			seenWg.Add(1)
			idx := i

			go func() {
				allSeenParts[idx] = observeSeen(seenSets)
				seenWg.Done()
			}()
		}

		workWg.Wait()

		close(errs)
		errsWg.Wait()

		close(seenSets)
		seenWg.Wait()

		allSeen := make([]int, 10000*19999)
		for _, part := range allSeenParts {
			for idx, id := range part {
				if id != 0 {
					allSeen[idx] = id
				}
			}
		}

		dictionaryBytes, err := ioutil.ReadFile(dictionaryFile)
		if err != nil {
			return err
		}
		dictionary := documents.FrequencyTable{}
		err = yaml.Unmarshal(dictionaryBytes, &dictionary)
		if err != nil {
			return err
		}

		numPrinted := 0
		for i := 0; i < 20000; i++ {
			toPrint := make(map[int]string)

			iWord := dictionary.Frequencies[i].Word
			if stopWords[iWord] {
				continue
			}

			iIdx := i * (i - 1) / 2
			for j := 0; j < i; j++ {
				jWord := dictionary.Frequencies[j].Word

				if stopWords[jWord] {
					continue
				}

				idx := iIdx + j
				if allSeen[idx] == 0 {
					toPrint[j] = jWord
				}
			}
			if len(toPrint) > 1000 {
				numPrinted++
				fmt.Println("Possible stop word:", i, iWord)
			} else if len(toPrint) > 0 {
				numPrinted += len(toPrint)
				nPrinted := 0
				for j, jWord := range toPrint {
					fmt.Printf("(%d, %d): (%s, %s)\n", i, j, iWord, jWord)
					nPrinted++
					if nPrinted >= 10 {
						continue
					}
				}
			}

			if numPrinted > 100 {
				break
			}
		}

		return nil
	},
}

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func readWordSets(path string) (*documents.WordSets, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	doc := &documents.WordSets{}
	err = json.Unmarshal(bytes, doc)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func observeSeen(wordSets <-chan *documents.WordSets) []int {
	result := make([]int, 10000*19999)

	for sets := range wordSets {
		for _, set := range sets.Documents {
			for _, i := range set.Words {
				if i == 0 {
					continue
				}
				i64 := int64(i)
				// Note that j < i
				iIdx := i64 * (i64 - 1) / 2
				for _, j := range set.Words {
					if j >= i {
						break
					}
					j64 := int64(j)

					idx := iIdx + j64

					if result[idx] != 0 {
						continue
					}

					result[idx] = set.ID
				}
			}
		}
	}

	return result
}
