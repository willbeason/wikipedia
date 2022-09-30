package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/willbeason/wikipedia/pkg/classify"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/nlp"
	"github.com/willbeason/wikipedia/pkg/ordinality"
	"github.com/willbeason/wikipedia/pkg/protos"
)

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			inDBPath := args[0]
			inTrainingData := args[1]
			inDictionary := args[2]
			articleIDsString := args[3]

			inDB, err := badger.Open(badger.DefaultOptions(inDBPath))
			if err != nil {
				return err
			}
			defer func() {
				_ = inDB.Close()
			}()

			parallel, err := cmd.Flags().GetInt(flags.ParallelKey)
			if err != nil {
				return err
			}

			errs, errsWg := jobs.Errors()

			known := &classify.ClassifiedArticles{}
			err = protos.Read(inTrainingData, known)
			fmt.Printf("Training on %d articles\n", len(known.Articles))

			if err != nil {
				return fmt.Errorf("unable to read training data: %w", err)
			}
			trainingPageIDs := make([]uint, len(known.Articles))

			i := 0
			for k := range known.Articles {
				trainingPageIDs[i] = uint(k)
				i++
			}

			trainingWork := jobs.IDs(inDB, newWordBag, trainingPageIDs, errs)
			trainingData := make(chan *classify.WordBagClassification, 100)
			trainingWorkWg := jobs.RunProto(parallel, readTrainingData(known.Articles, trainingData), trainingWork, errs)

			dictionary, err := nlp.ReadDictionary(inDictionary)
			if err != nil {
				return err
			}

			modelChan := classify.TrainBayes(len(classify.Classification_value)-1, len(dictionary.Words), trainingData)

			trainingWorkWg.Wait()
			close(trainingData)

			model := <-modelChan
			fmt.Println("Trained model")

			var ids []uint
			for _, articleIDString := range strings.Split(articleIDsString, ",") {
				articleID, err2 := strconv.ParseUint(articleIDString, 10, 32)
				if err2 != nil {
					return fmt.Errorf("article ID %s is not a valid uint32", articleIDString)
				}
				ids = append(ids, uint(articleID))
			}

			findWork := jobs.IDs(inDB, newWordBag, trainingPageIDs, errs)
			foundChan := make(chan *ordinality.PageWordBag, 100)
			findWorkWg := jobs.RunProto(parallel, findPage(foundChan), findWork, errs)

			classifyWg := sync.WaitGroup{}
			classifyWg.Add(1)

			wrong := 0
			go func() {
				for found := range foundChan {
					result := model.Classify(found)
					if result[0].Classification != known.Articles[found.Id] {
						wrong++
						fmt.Println()
						fmt.Printf("Incorrectly classified %q as %s, want %s\n", found.Title, result[0].Classification, known.Articles[found.Id])
						for _, cp := range result[:4] {
							fmt.Printf("%s: %.4f%%\n", cp.Classification.String(), cp.P*100)
						}
						fmt.Println()
					}
				}

				classifyWg.Done()
			}()

			findWorkWg.Wait()

			err = inDB.Close()
			if err != nil {
				return err
			}

			close(foundChan)
			close(errs)
			errsWg.Wait()

			classifyWg.Wait()

			fmt.Printf("Accuracy: %.2f%%", 100-float64(wrong)*100/float64(len(known.Articles)))

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

func readTrainingData(known map[uint32]classify.Classification, trainingData chan<- *classify.WordBagClassification) jobs.Proto {
	return func(p proto.Message) error {
		wordBag, ok := p.(*ordinality.PageWordBag)
		if !ok {
			return fmt.Errorf("got proto.Message type %T, want %T", p, &ordinality.PageWordBag{})
		}

		classification := known[wordBag.Id]
		if classification == classify.Classification_UNKNOWN {
			return fmt.Errorf("unknown classification for article ID %d", wordBag.Id)
		}

		trainingData <- &classify.WordBagClassification{
			Classification: classification,
			PageWordBag:    wordBag,
		}

		return nil
	}
}

func findPage(found chan<- *ordinality.PageWordBag) jobs.Proto {
	return func(proto proto.Message) error {
		page, ok := proto.(*ordinality.PageWordBag)
		if !ok {
			return fmt.Errorf("got proto.Message type %T, want %T", proto, &ordinality.PageWordBag{})
		}

		found <- page

		return nil
	}
}
