package gender

import (
	"context"
	"errors"
	"fmt"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/config"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/entities"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/pages"
	"github.com/willbeason/wikipedia/pkg/protos"
)

var ErrIndex = errors.New("indexing biography genders")

func Index(cmd *cobra.Command, cfg *config.GenderIndex, corpusNames ...string) error {
	if len(corpusNames) != 1 {
		return fmt.Errorf("%w: must have exactly one corpus but got %+v", ErrGenderFrequency, corpusNames)
	}
	corpusName := corpusNames[0]

	parallel, err := flags.GetParallel(cmd)
	if err != nil {
		return err
	}

	workspace, err := flags.GetWorkspacePath(cmd)
	if err != nil {
		return err
	}

	wikidataDB := filepath.Join(workspace, corpusName, cfg.Wikidata)
	source := pages.StreamDB[entities.Entity](wikidataDB, parallel)

	ctx, cancel := context.WithCancelCause(cmd.Context())
	entitiesChan, err := source(ctx, cancel)
	if err != nil {
		return err
	}

	protosChan := make(chan *documents.ArticleIdGender, 10000)
	transformWg := &sync.WaitGroup{}

	for range parallel {
		transformWg.Add(1)

		go func() {
		EntitiesLoop:
			for {
				select {
				case <-ctx.Done():
					break EntitiesLoop
				case entity, ok := <-entitiesChan:
					if !ok {
						break EntitiesLoop
					}

					inferredGender, inferErr := processEntity(entity)
					switch {
					case errors.Is(inferErr, ErrNotHuman):
						// Ignore non-humans.
					case inferErr != nil:
						cancel(inferErr)
						break EntitiesLoop
					default:
						protosChan <- &documents.ArticleIdGender{
							Id:     entity.Id,
							Gender: inferredGender,
						}
					}
				}
			}

			transformWg.Done()
		}()
	}

	go func() {
		transformWg.Wait()
		close(protosChan)
	}()

	errs := make(chan error)
	go func() {
		for err := range errs {
			cancel(err)
		}
	}()

	outFile := filepath.Join(workspace, corpusName, cfg.Out)
	writeSink := jobs.NewSink(protos.WriteFile[*documents.ArticleIdGender](outFile))
	writeSinkWg, writeSinkJob := writeSink(protosChan)
	go writeSinkJob(ctx, errs)

	writeSinkWg.Wait()

	err = ctx.Err()
	if err != nil {
		return fmt.Errorf("%w: writing gender index: %w", ErrIndex, err)
	}

	return nil
}
