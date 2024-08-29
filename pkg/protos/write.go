package protos

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/willbeason/wikipedia/pkg/jobs"
	"google.golang.org/protobuf/proto"
)

func WriteFile[OUT proto.Message](path string) jobs.SinkFn[OUT] {
	return func(out <-chan OUT) jobs.Job {
		return writeFileJob[OUT](path, out)
	}
}

func writeFileJob[OUT proto.Message](path string, out <-chan OUT) jobs.Job {
	return func(ctx context.Context, errs chan<- error) {
		writeStream[OUT](ctx, errs, path, out)
	}
}

func writeStream[OUT proto.Message](ctx context.Context, errs chan<- error, path string, out <-chan OUT) {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		errs <- err
		return
	}

	file, err := os.Create(path)
	if err != nil {
		errs <- err
		return
	}

	defer closeFile(file, errs)

	writer := bufio.NewWriter(file)

WriteLoop:
	for {
		select {
		case <-ctx.Done():
			break WriteLoop
		case p, ok := <-out:
			if !ok {
				// Protos channel is closed.
				err = writer.Flush()
				if err != nil {
					errs <- fmt.Errorf("flushing %q: %w", path, err)
				}
				break WriteLoop
			}

			err = writeNextMessage(writer, p)
			if err != nil {
				errs <- err
				break WriteLoop
			}
		}
	}
}

func writeNextMessage(writer io.Writer, p proto.Message) error {
	bytes, marshalErr := proto.Marshal(p)
	if marshalErr != nil {
		return fmt.Errorf("marhsaling %T: %w", p, marshalErr)
	}

	// Record length of marshaled message.
	lengthBuf := make([]byte, SizeLen)
	binary.LittleEndian.PutUint32(lengthBuf, uint32(len(bytes)))
	if _, writeErr := writer.Write(lengthBuf); writeErr != nil {
		return fmt.Errorf("writing size of next message: %w", writeErr)
	}

	// Record message.
	if _, writeErr := writer.Write(bytes); writeErr != nil {
		return fmt.Errorf("writing next message bytes: %w", writeErr)
	}

	return nil
}
