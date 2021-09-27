package main

import (
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/classify"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/nlp"
	"github.com/willbeason/wikipedia/pkg/ordinality"
	"github.com/willbeason/wikipedia/pkg/protos"
	"google.golang.org/protobuf/proto"
)

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			inDBPath := args[0]
			inTrainingData := args[1]
			inDictionary := args[2]

			inDB, err := badger.Open(badger.DefaultOptions(inDBPath))
			if err != nil {
				return err
			}
			defer func() {
				_ = inDB.Close()
			}()

			errs, errsWg := jobs.Errors()

			known := &classify.ClassifiedArticles{}
			err = protos.Read(inTrainingData, known)
			if err != nil {
				return fmt.Errorf("unable to read training data: %w", err)
			}

			work := jobs.IDs(inDB, newWordBag, known.ToIDs(), errs)

			dictionary, err := nlp.ReadDictionary(inDictionary)
			if err != nil {
				return err
			}

			// printWg := collect(known.Articles, work, errs)
			printWg := findMissing(dictionary, work, errs)

			printWg.Wait()
			close(errs)

			errsWg.Wait()

			err = inDB.Close()
			if err != nil {
				return err
			}

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

func newWordBag() proto.Message {
	return &ordinality.PageWordBag{}
}

func findMissing(dictionary *nlp.Dictionary, in <-chan proto.Message, errs chan<- error) *sync.WaitGroup {
	counts := make([]idCount, len(dictionary.Words))
	for i := range counts {
		counts[i].id = i + 1
		counts[i].word = dictionary.Words[i]
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for p := range in {
			wordBag, ok := p.(*ordinality.PageWordBag)
			if !ok {
				errs <- fmt.Errorf("got proto.Message type %T, want %T", p, &ordinality.PageWordBag{})
				continue
			}

			for _, wc := range wordBag.Words {
				counts[wc.Word-1].count += wc.Count
			}
		}
		counts = counts[:5000]

		sort.Slice(counts, func(i, j int) bool {
			if counts[i].count != counts[j].count {
				return counts[i].count < counts[j].count
			}
			return counts[i].id < counts[j].id
		})

		for _, wc := range counts {
			if wc.count > 9 {
				break
			}

			fmt.Printf("%d: %q: %d\n", wc.id, wc.word, wc.count)
		}

		wg.Done()
	}()

	return &wg
}

type idCount struct {
	id    int
	count uint32
	word  string
}

func collect(known map[uint32]classify.Classification, in <-chan proto.Message, errs chan<- error) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		result := make([]classificationCount, len(classify.Classification_value))
		for i := range result {
			result[i].Classification = classify.Classification(i)
		}

		for p := range in {
			wordBag, ok := p.(*ordinality.PageWordBag)
			if !ok {
				errs <- fmt.Errorf("got proto.Message type %T, want %T", p, &ordinality.PageWordBag{})
				continue
			}

			result[known[wordBag.Id]].count += getSize(wordBag)
		}

		sort.Slice(result, func(i, j int) bool {
			return result[i].count < result[j].count
		})

		for _, r := range result {
			fmt.Println(r.Classification.String(), ":", r.count)
		}

		wg.Done()
	}()

	return wg
}

func getSize(wordBag *ordinality.PageWordBag) uint32 {
	size := uint32(0)

	for _, wc := range wordBag.Words {
		if wc.Word < classify.Ignored {
			continue
		}

		size += wc.Count
	}

	return size
}

type classificationCount struct {
	classify.Classification
	count uint32
}
