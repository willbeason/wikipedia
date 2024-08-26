package protos

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/willbeason/wikipedia/pkg/jobs"
)

// WriteDir writes a stream of uniquely-identified protos to a directory.
// Shards protos into the number file shards specified.
func WriteDir[OUT ID](dir string, shards uint32) jobs.SinkFn[OUT] {
	return func(out <-chan OUT) jobs.Job {
		return writeDirJob(dir, shards, out)
	}
}

func writeDirJob[OUT ID](dir string, shards uint32, out <-chan OUT) jobs.Job {
	return func(ctx context.Context, errs chan<- error) {
		writeDir(ctx, errs, dir, shards, out)
	}
}

func writeDir[OUT ID](ctx context.Context, errs chan<- error, dir string, shards uint32, out <-chan OUT) {
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		errs <- err
		return
	}

	shardChannels := make([]chan OUT, shards)

	// Initialize WaitGroup so that we ensure all goroutines exit before leaving writeDir.
	wg := &sync.WaitGroup{}
	// Ensure all shard channels are closed upon cancelled context.
	wg.Add(1)

	for i := range shards {
		// Delta to ensure each shard finishes all writing work.
		wg.Add(1)

		filename := fmt.Sprintf("shard_%03d.pdb", i)
		path := filepath.Join(dir, filename)

		shardChannel := make(chan OUT)
		// Per-shard go function writing to an individual shard.
		go func() {
			defer wg.Done()
			writeStream(ctx, errs, path, shardChannel)
		}()
		shardChannels[i] = shardChannel
	}

	// Router go function that picks which shard to write each proto.
	go route(ctx, wg, shardChannels, out)

	wg.Wait()
}

func route[OUT ID](ctx context.Context, wg *sync.WaitGroup, shardChannels []chan OUT, out <-chan OUT) {
	defer func() {
		for _, shardChannel := range shardChannels {
			close(shardChannel)
		}
		wg.Done()
	}()

	shards := uint32(len(shardChannels))
	for {
		select {
		case <-ctx.Done():
			return
		case p, ok := <-out:
			if !ok {
				// No more protos to write.
				return
			}

			// Assign the proto to a shard based on its identifier and pass it
			// to the appropriate channel.
			shard := p.ID() % shards
			select {
			case <-ctx.Done():
				// Shard channels may no longer be being read from if context is cancelled.
				return
			case shardChannels[shard] <- p:
				// Normal case.
			}
		}
	}
}
