package main

import (
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/classify"
	"github.com/willbeason/extract-wikipedia/pkg/flags"
	"github.com/willbeason/extract-wikipedia/pkg/jobs"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"
	"github.com/willbeason/extract-wikipedia/pkg/ordinality"
	"google.golang.org/protobuf/proto"
)

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			inWordBags := args[0]
			inDictionary := args[1]
			articleIDString := args[2]
			articleID, err := strconv.ParseInt(articleIDString, 10, 64)
			if err != nil {
				return fmt.Errorf("article ID %s is not a valid integer", articleIDString)
			}
			if articleID > math.MaxUint32 {
				panic(articleIDString)
			}

			parallel, err := cmd.Flags().GetInt(flags.ParallelKey)
			if err != nil {
				return err
			}

			dictionary, err := nlp.ReadDictionary(inDictionary)

			errs, errsWg := jobs.Errors()

			trainingWork := jobs.WalkFiles(inWordBags, errs)

			trainingData := make(chan *classify.WordBagClassification, 100)

			known := classify.Base()
			trainingWorkWg := jobs.RunProtoJobs(parallel, newDocumentWordBag, readTrainingData(known, trainingData), trainingWork, errs)

			modelChan := classify.TrainBayes(len(classify.Classification_value)-1, len(dictionary.Words), trainingData)

			trainingWorkWg.Wait()
			close(trainingData)

			model := <- modelChan
			fmt.Println("Trained model")

			findWork := jobs.WalkFiles(inWordBags, errs)
			foundChan := make(chan *ordinality.PageWordBag, 1)
			findWorkWg := jobs.RunProtoJobs(parallel, newDocumentWordBag, findPage(uint32(articleID), foundChan), findWork, errs)

			findWorkWg.Wait()
			close(foundChan)

			close(errs)
			errsWg.Wait()

			found := <- foundChan

			result := model.Classify(found)

			for _, cp := range result {
				fmt.Printf("%s: %.4f%%\n", cp.Classification.String(), cp.P*100)
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

func newDocumentWordBag() proto.Message {
	return &ordinality.DocumentWordBag{}
}

func readTrainingData(known *classify.ClassifiedArticles, trainingData chan<- *classify.WordBagClassification) jobs.Proto {
	return func(proto proto.Message) error {
		doc, ok := proto.(*ordinality.DocumentWordBag)
		if !ok {
			return fmt.Errorf("got proto.Message type %T, want %T", proto, &ordinality.DocumentWordBag{})
		}

		for _, page := range doc.Pages {
			classification := known.Articles[page.Id]
			if classification == classify.Classification_UNKNOWN {
				continue
			}

			fmt.Printf("Found article %d: %q to classify as %s\n", page.Id, page.Title, classification)
			trainingData <- &classify.WordBagClassification{
				Classification: classification,
				PageWordBag:    page,
			}
		}

		return nil
	}
}

func findPage(want uint32, found chan<- *ordinality.PageWordBag) jobs.Proto {
	return func(proto proto.Message) error {
		doc, ok := proto.(*ordinality.DocumentWordBag)
		if !ok {
			return fmt.Errorf("got proto.Message type %T, want %T", proto, &ordinality.DocumentWordBag{})
		}

		for _, page := range doc.Pages {
			if page.Id == want {
				fmt.Printf("Found article %d: %q\n", want, page.Title)
				found <- page
			}
		}

		return nil
	}
}
