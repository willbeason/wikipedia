package gender

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/willbeason/wikipedia/pkg/protos"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/config"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/entities"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/pages"
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
	ps, err := source(ctx, cancel)
	if err != nil {
		return err
	}

	articleGendersChan := jobs.Reduce2[*entities.Entity, map[uint32]string](
		ctx,
		cancel,
		parallel,
		10000,
		ps,
		jobs.NewMap[uint32, string],
		genderMapWorkFn,
		jobs.MergeInto[uint32, string],
	)

	articleGenders := <-articleGendersChan

	result := &documents.GenderIndex{
		Genders: make(map[uint32]string),
	}
	for id, gender := range articleGenders {
		result.Genders[id] = gender
	}

	outFile := filepath.Join(workspace, corpusName, cfg.Out)
	err = protos.Write(outFile, result)
	if err != nil {
		return fmt.Errorf("%w: writing gender index: %w", ErrIndex, err)
	}

	return nil
}
