package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/willbeason/extract-wikipedia/pkg/flags"

	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/walker"
	"gopkg.in/yaml.v3"
)

const (
	MaxPrint          = 100
	StopWordThreshold = 1000
)

func StopWords() map[string]bool {
	return map[string]bool{
		"noinclude":      true,
		"afd":            true,
		"wikiproject":    true,
		"diff":           true,
		"xfd":            true,
		"boilerplate":    true,
		"overlap":        true,
		"notability":     true,
		"em":             true,
		"php":            true,
		"blacklist":      true,
		"whitelist":      true,
		"td":             true,
		"edit":           true,
		"disambiguation": true,
		"username":       true,
		"admin":          true,
		"metadata":       true,
		"bot":            true,
		"edits":          true,
		"tr":             true,
		"templates":      true,
		"uploaded":       true,
		"utc":            true,
		"whatlinkshere":  true,
		"gng":            true,
		"delete":         true,
		"overlaps":       true,
		"tt":             true,
		"deletion":       true,
		"coi":            true,
		"stub":           true,
		"revert":         true,
		"nominations":    true,
		"afc":            true,
		"sorting":        true,
		"speedy":         true,
		"redirect":       true,
		"user":           true,
		"https":          true,
		"http":           true,
		"margin":         true,
		"nom":            true,
		"mediawiki":      true,
		"lx":             true,
		"posted":         true,
		"manually":       true,
		"middot":         true,
		"vandalism":      true,
		"rfc":            true,
		"nominate":       true,
		"logs":           true,
		"unsigned":       true,
		"pov":            true,
		"assessed":       true,
		"upload":         true,
		"htm":            true,
		"gt":             true,
		"navbox":         true,
		"frac":           true,
		"uefa":           true,
		"notify":         true,
		"contribution":   true,
		"padding":        true,
		"blp":            true,
		"tbl":            true,
		"rfd":            true,
		"admins":         true,
		"deleting":       true,
		"checkuser":      true,
	}
}

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			dictionarySize, err := cmd.Flags().GetInt(flags.DictionarySizeKey)
			if err != nil {
				return err
			}

			nParallel, err := cmd.Flags().GetInt(flags.ParallelKey)
			if err != nil {
				return err
			}

			wordSetsDir := args[0]
			dictionaryFile := args[1]

			work := make(chan string)
			errs := make(chan error)

			go func() {
				err2 := filepath.WalkDir(wordSetsDir, walker.Files(work))
				if err2 != nil {
					errs <- err2
				}
				close(work)
			}()

			workWg := sync.WaitGroup{}

			seenSets := make(chan *documents.WordSets)

			for i := 0; i < nParallel; i++ {
				workWg.Add(1)
				go func() {
					for item := range work {
						newWordSets, err2 := readWordSets(item)
						if err2 != nil {
							errs <- fmt.Errorf("%s: %w", item, err2)
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

			allSeenParts := make([][]int, nParallel)
			seenWg := sync.WaitGroup{}

			for i := 0; i < nParallel; i++ {
				seenWg.Add(1)
				idx := i

				go func() {
					allSeenParts[idx] = observeSeen(dictionarySize, seenSets)
					seenWg.Done()
				}()
			}

			workWg.Wait()

			close(errs)
			errsWg.Wait()

			close(seenSets)
			seenWg.Wait()

			allSeen := collectSeen(dictionarySize, allSeenParts)

			dictionaryBytes, err := ioutil.ReadFile(dictionaryFile)
			if err != nil {
				return err
			}
			dictionary := documents.FrequencyTable{}
			err = yaml.Unmarshal(dictionaryBytes, &dictionary)
			if err != nil {
				return err
			}

			printResults(allSeen, dictionary)

			return nil
		},
	}

	flags.Parallel(cmd)

	return cmd
}

func main() {
	err := mainCmd().Execute()
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

func observeSeen(dictionarySize int, wordSets <-chan *documents.WordSets) []int {
	result := make([]int, dictionarySize*(dictionarySize-1)/2)

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

					switch result[idx] {
					case 0:
						// First seen
						result[idx] = set.ID
					case -1:
						// Do nothing
					default:
						result[idx] = -1
					}
				}
			}
		}
	}

	return result
}

func collectSeen(dictionarySize int, allSeenParts [][]int) []int {
	allSeen := make([]int, dictionarySize*(dictionarySize-1)/2)

	for _, part := range allSeenParts {
		for idx, id := range part {
			switch id {
			case 0:
				// The pair was not observed in this slice.
			case -1:
				// The pair was observed multiple times in this slice.
				allSeen[idx] = -1
			default:
				switch allSeen[idx] {
				case 0:
					// The pair has not been seen before, and was seen once.
					allSeen[idx] = id
				case -1:
					// The pair has already been seen multiple times.
				default:
					// The pair has been seen once before, and now once more.
					allSeen[idx] = -1
				}
			}
		}
	}

	return allSeen
}

func printResults(allSeen []int, dictionary documents.FrequencyTable) {
	numPrinted := 0
	alreadyPrinted := make(map[int]int)
	stopWords := StopWords()

	for i := 0; i < 20000; i++ {
		toPrint := make(map[int]string)

		if alreadyPrinted[i] != 0 {
			continue
		}

		iWord := dictionary.Frequencies[i].Word
		if stopWords[iWord] {
			continue
		}

		iIdx := i * (i - 1) / 2

		for j := 0; j < i; j++ {
			if alreadyPrinted[j] != 0 || alreadyPrinted[i] != 0 {
				continue
			}

			jWord := dictionary.Frequencies[j].Word

			if stopWords[jWord] {
				continue
			}

			idx := iIdx + j
			switch allSeen[idx] {
			case 0, -1:
				// Do nothing
			default:
				// Seen exactly once
				alreadyPrinted[i] = allSeen[idx]
				alreadyPrinted[j] = allSeen[idx]
				toPrint[j] = jWord
			}
		}

		numPrinted += printResult(i, iWord, toPrint, alreadyPrinted)

		if numPrinted > MaxPrint {
			break
		}
	}
}

func printResult(i int, iWord string, toPrint map[int]string, alreadyPrinted map[int]int) int {
	if len(toPrint) > StopWordThreshold {
		fmt.Println("Possible stop word:", i, iWord)
		return 1
	}

	nPrinted := 0

	if len(toPrint) > 0 {
		for j, jWord := range toPrint {
			fmt.Printf("(%d, %d): (%s, %s): %d\n", i, j, iWord, jWord, alreadyPrinted[i])
			nPrinted++

			if nPrinted >= 10 {
				break
			}
		}
	}

	return nPrinted
}
