package documents

import (
	"context"
	"errors"
	"fmt"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/protos"
)

var ErrLoop = errors.New("redirect loop detected")


func MakeRedirects(ctx context.Context, filename string, errs chan<- error) <-chan map[string]string {
	redirectSource := jobs.NewSource(protos.ReadFile[Redirect](filename))
	_, redirectsJob, redirects := redirectSource()
	go redirectsJob(ctx, errs)

	redirectsReduce := jobs.NewMap(MakeRedirectsMapFn)
	_, redirectsReduceJob, futureIndexes := redirectsReduce(redirects)
	go redirectsReduceJob(ctx, errs)

	return futureIndexes
}

func MakeRedirectsMapFn(redirects <-chan *Redirect, redirectMap chan<- map[string]string) jobs.Job {
	return func(ctx context.Context, _ chan<- error) {
		result := make(map[string]string)
		defer func() {
			redirectMap <- result
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case redirect, ok := <-redirects:
				if !ok {
					return
				}

				result[redirect.Title] = redirect.Redirect
			}
		}
	}
}

func GetDestination(redirects map[string]string, titleIndex map[string]uint32, title string) (string, error) {
	if _, notRedirect := titleIndex[title]; notRedirect {
		// This isn't actually a redirect since the article really exists.
		return title, nil
	}

	destination, isRedirect := redirects[title]
	if title == destination {
		// Self redirect.
		return title, nil
	}

	seen := map[string]bool{
		title: true,
	}
	loop := []string{title}

	for ; isRedirect; destination, isRedirect = redirects[title] {
		loop = append(loop, destination)

		if seen[destination] {
			return "", fmt.Errorf("%w: %v", ErrLoop, loop)
		}

		seen[destination] = true

		title = destination
		if _, notRedirect := titleIndex[title]; notRedirect {
			// This isn't actually a redirect since the article really exists.
			return title, nil
		}
	}

	return title, nil
}
