package documents

import (
	"context"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/protos"
)

func ReadTitleMap(ctx context.Context, filename string, errs chan<- error) <-chan map[string]uint32 {
	titlesSource := jobs.NewSource(protos.ReadFile[ArticleIdTitle](filename))
	_, titlesJob, titleIds := titlesSource()
	go titlesJob(ctx, errs)

	titleReduce := jobs.NewMap(jobs.Reduce3(jobs.MakeMap[string, uint32], mergeTitleMap))
	_, titleReduceJob, futureIndex := titleReduce(titleIds)
	go titleReduceJob(ctx, errs)

	return futureIndex
}

func mergeTitleMap(title *ArticleIdTitle, titleMap map[string]uint32) error {
	titleMap[title.Title] = title.Id
	return nil
}

func ReadGenderMap(ctx context.Context, filename string, errs chan<- error) <-chan map[uint32]string {
	genderSource := jobs.NewSource(protos.ReadFile[ArticleIdGender](filename))
	_, genderJob, genderIds := genderSource()
	go genderJob(ctx, errs)

	genderReduce := jobs.NewMap(jobs.Reduce3(jobs.MakeMap[uint32, string], mergeGenderMap))
	_, genderReduceJob, futureIndex := genderReduce(genderIds)
	go genderReduceJob(ctx, errs)

	return futureIndex
}

func mergeGenderMap(gender *ArticleIdGender, genderMap map[uint32]string) error {
	genderMap[gender.Id] = gender.Gender
	return nil
}
