package gender

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"sync"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/config"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
)

var ErrGenderFrequency = errors.New("running gender frequency calculation")

const Claim = "P21"

type KeyCount struct {
	Key   string
	Count int
}

func Frequency(cmd *cobra.Command, cfg *config.GenderFrequency, corpusNames ...string) error {
	if len(corpusNames) != 1 {
		return fmt.Errorf("%w: must have exactly one corpus but got %+v", ErrGenderFrequency, corpusNames)
	}
	corpusName := corpusNames[0]

	workspace, err := flags.GetWorkspacePath(cmd)
	if err != nil {
		return err
	}

	genderIndexFile := filepath.Join(workspace, corpusName, cfg.GenderIndex)

	ctx, cancel := context.WithCancelCause(cmd.Context())

	errs := make(chan error, 1)
	errsWg := sync.WaitGroup{}
	errsWg.Add(1)
	go func() {
		for err := range errs {
			cancel(err)
		}
		errsWg.Done()
	}()

	articleGendersFuture := documents.ReadGenderMap(ctx, genderIndexFile, errs)
	articleGenders := <-articleGendersFuture

	fmt.Println(len(articleGenders))
	result := make(map[string]int)
	for _, gender := range articleGenders {
		result[gender]++
	}

	counts := make([]KeyCount, 0, len(result))
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

	close(errs)
	errsWg.Wait()
	if ctx.Err() != nil {
		return context.Cause(ctx)
	}

	return nil
}
