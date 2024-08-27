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

	titleReduce := jobs.NewMap(jobs.ReduceToOne(jobs.MakeMap[string, uint32], mergeTitleMap))
	_, titleReduceJob, futureIndex := titleReduce(titleIds)
	go titleReduceJob(ctx, errs)

	return futureIndex
}

func mergeTitleMap(title *ArticleIdTitle, titleMap map[string]uint32) error {
	titleMap[title.Title] = title.Id
	return nil
}

func ReadReverseTitleMap(ctx context.Context, filename string, errs chan<- error) <-chan map[uint32]string {
	titlesSource := jobs.NewSource(protos.ReadFile[ArticleIdTitle](filename))
	_, titlesJob, titleIds := titlesSource()
	go titlesJob(ctx, errs)

	titleReduce := jobs.NewMap(jobs.ReduceToOne(jobs.MakeMap[uint32, string], mergeReverseTitleMap))
	_, titleReduceJob, futureIndex := titleReduce(titleIds)
	go titleReduceJob(ctx, errs)

	return futureIndex
}

func mergeReverseTitleMap(title *ArticleIdTitle, titleMap map[uint32]string) error {
	titleMap[title.Id] = title.Title
	return nil
}

func ReadGenderMap(ctx context.Context, filename string, errs chan<- error) <-chan map[uint32]string {
	genderSource := jobs.NewSource(protos.ReadFile[ArticleIdGender](filename))
	_, genderJob, genderIds := genderSource()
	go genderJob(ctx, errs)

	genderReduce := jobs.NewMap(jobs.ReduceToOne(jobs.MakeMap[uint32, string], mergeGenderMap))
	_, genderReduceJob, futureIndex := genderReduce(genderIds)
	go genderReduceJob(ctx, errs)

	return futureIndex
}

func mergeGenderMap(gender *ArticleIdGender, genderMap map[uint32]string) error {
	genderMap[gender.Id] = gender.Gender
	return nil
}

func ReadLinksMap(ctx context.Context, filename string, errs chan<- error) <-chan map[uint32][]uint32 {
	linkSource := jobs.NewSource(protos.ReadFile[ArticleIdLinks](filename))
	_, linkJob, linkIds := linkSource()
	go linkJob(ctx, errs)

	linkReduce := jobs.NewMap(jobs.ReduceToOne(jobs.MakeMap[uint32, []uint32], mergeArticleLinks))
	_, linksReduceJob, futureLinks := linkReduce(linkIds)
	go linksReduceJob(ctx, errs)

	return futureLinks
}

func mergeArticleLinks(links *ArticleIdLinks, linkMap map[uint32][]uint32) error {
	result := make([]uint32, len(links.Links))
	for i, link := range links.Links {
		result[i] = link.Target
	}
	linkMap[links.Id] = result

	return nil
}
