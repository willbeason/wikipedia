package gender

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/config"
	"github.com/willbeason/wikipedia/pkg/entities"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/pages"
)

func Comparison(cmd *cobra.Command, cfg *config.GenderComparison, corpusNames ...string) error {
	// if len(corpusNames) != 2 {
	//	return fmt.Errorf("%w: must have exactly two corpora but got %+v", ErrGenderFrequency, corpusNames)
	//}
	// beforeCorpusName := corpusNames[0]
	//afterCorpusName := corpusNames[1]
	//
	//parallel, err := flags.GetParallel(cmd)
	//if err != nil {
	//	return err
	//}
	//
	//workspace, err := flags.GetWorkspacePath(cmd)
	//if err != nil {
	//	return err
	//}

	return nil
}

func getArticleGenders(ctx context.Context, cancel context.CancelCauseFunc, parallel int, wikidataDB string) (map[uint32]string, error) {
	source := pages.StreamDB[entities.Entity](wikidataDB, parallel)

	ps, err := source(ctx, cancel)
	if err != nil {
		return nil, err
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

	return <-articleGendersChan, nil
}
