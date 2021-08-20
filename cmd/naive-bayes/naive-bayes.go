package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/classify"
	"github.com/willbeason/extract-wikipedia/pkg/flags"
	"github.com/willbeason/extract-wikipedia/pkg/jobs"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"
	"github.com/willbeason/extract-wikipedia/pkg/ordinality"
	"github.com/willbeason/extract-wikipedia/pkg/protos"
	"google.golang.org/protobuf/proto"
)

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			inDBPath := args[0]
			inTrainingData := args[1]
			inDictionary := args[2]
			articleIDString := args[3]

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
			if err != nil {
				return fmt.Errorf("unable to read training data: %w", err)
			}
			trainingPageIds := make([]uint, len(known.Articles))

			i := 0
			for k := range known.Articles {
				trainingPageIds[i] = uint(k)
				i++
			}

			trainingWork := jobs.IDs(inDB, newWordBag, trainingPageIds, errs)
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

			articleID, err := strconv.ParseUint(articleIDString, 10, 32)
			if err != nil {
				return fmt.Errorf("article ID %s is not a valid uint32", articleIDString)
			}

			findWork := jobs.IDs(inDB, newWordBag, []uint{uint(articleID)}, errs)
			foundChan := make(chan *ordinality.PageWordBag, 100)
			findWorkWg := jobs.RunProto(parallel, findPage(foundChan), findWork, errs)

			findWorkWg.Wait()

			found := <-foundChan

			err = inDB.Close()
			if err != nil {
				return err
			}

			close(foundChan)
			close(errs)
			errsWg.Wait()

			result := model.Classify(found)
			fmt.Println()
			fmt.Printf("Classifying article %q\n", found.Title)
			for _, cp := range result {
				fmt.Printf("%s: %.4f%%\n", cp.Classification.String(), cp.P*100)
			}
			fmt.Println()

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
	return func(proto proto.Message) error {
		page, ok := proto.(*ordinality.PageWordBag)
		if !ok {
			return fmt.Errorf("got proto.Message type %T, want %T", proto, &ordinality.PageWordBag{})
		}

		classification := known[page.Id]
		if classification == classify.Classification_UNKNOWN {
			return fmt.Errorf("unknown classification for article ID %d", page.Id)
		}

		fmt.Printf("Found article %d: %q to classify as %s\n", page.Id, page.Title, classification)
		trainingData <- &classify.WordBagClassification{
			Classification: classification,
			PageWordBag:    page,
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
