package gender_frequency

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"

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

	genderMaps := make(chan map[string]int)
	resultChan := make(chan map[string]int)
	var noClaims atomic.Int64
	var multipleClaims atomic.Int64

	entityWg := sync.WaitGroup{}
	genderWg := sync.WaitGroup{}
	for range parallel {
		entityWg.Add(1)
		genderWg.Add(1)
		go func() {
			genderMap := make(map[string]int)

			for entity := range ps {
				instanceOfClaims := entity.Claims[ingest_wikidata.InstanceOf]
				isHuman := false
				for _, claim := range instanceOfClaims.Claim {
					if claim.Value == "Q5" {
						isHuman = true
					}
				}
				if !isHuman {
					continue
				}

				genderClaims, hasGenderClaims := entity.Claims[GenderClaim]
				if !hasGenderClaims {
					noClaims.Add(1)
					continue
				}
				claims := genderClaims.Claim
				switch len(claims) {
				case 0:
					noClaims.Add(1)
				case 1:
					claim := claims[0]
					if claim.Value == "Q301702" {
						fmt.Println(entity.Sitelinks["enwiki"].Title)
					}
					genderMap[claim.Value]++
				default:
					multipleClaims.Add(1)
				}
			}

			entityWg.Done()

			genderMaps <- genderMap
			genderWg.Done()
		}()
	}

	go func() {
		result := make(map[string]int)
		for m := range genderMaps {
			for k, v := range m {
				result[k] += v
			}
		}
		resultChan <- result
		close(resultChan)
	}()

	entityWg.Wait()
	genderWg.Wait()
	close(genderMaps)

	result := <-resultChan

	var counts []KeyCount
	for k, v := range result {
		counts = append(counts, KeyCount{k, v})
	}
	sort.Slice(counts, func(i, j int) bool {
		return counts[i].Count > counts[j].Count
	})

	for _, c := range counts {
		fmt.Printf("%s, %d\n", c.Key, c.Count)
	}
	fmt.Println("no claims", noClaims.Load())
	fmt.Println("multiple claims", multipleClaims.Load())

	return nil
}
