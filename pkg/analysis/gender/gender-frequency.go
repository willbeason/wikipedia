package gender

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/willbeason/wikipedia/pkg/jobs"

	ingest_wikidata "github.com/willbeason/wikipedia/pkg/ingest-wikidata"

	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/pages"

	"github.com/willbeason/wikipedia/pkg/config"
	"github.com/willbeason/wikipedia/pkg/entities"

	"github.com/spf13/cobra"
)

var ErrGenderFrequency = errors.New("running gender frequency calculation")

const GenderClaim = "P21"

type KeyCount struct {
	Key   string
	Count int
}

func GenderFrequency(cmd *cobra.Command, cfg *config.GenderFrequency, corpusNames ...string) error {
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

	wikidataDB := filepath.Join(workspace, corpusName, cfg.In)
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

	result := make(map[string]int)
	for _, gender := range articleGenders {
		result[gender]++
	}

	var counts []KeyCount
	for k, v := range result {
		counts = append(counts, KeyCount{k, v})
	}
	sort.Slice(counts, func(i, j int) bool {
		if counts[i].Count != counts[j].Count {
			return counts[i].Count > counts[j].Count
		}
		return counts[i].Key < counts[j].Key
	})

	for _, c := range counts {
		fmt.Printf("%s, %d\n", c.Key, c.Count)
	}

	return nil
}

func genderMapWorkFn(out map[uint32]string, e *entities.Entity) error {
	inferredGender, err := processEntity(e)

	switch {
	case errors.Is(err, ErrNotHuman):
		// Ignore
	case err != nil:
		return err
	default:
		out[e.Id] = inferredGender
	}

	return nil
}

var ErrNotHuman = errors.New("entity is not human")

func processEntity(entity *entities.Entity) (string, error) {
	instanceOfClaims := entity.Claims[ingest_wikidata.InstanceOf]
	isHuman := false
	for _, claim := range instanceOfClaims.Claim {
		if claim.Value == "Q5" {
			isHuman = true
		}
	}
	if !isHuman {
		return "", ErrNotHuman
	}

	genderClaims, hasGenderClaims := entity.Claims[GenderClaim]
	if !hasGenderClaims {
		return NoClaims, nil
	}
	claims := genderClaims.Claim
	return Infer(claims), nil
}
