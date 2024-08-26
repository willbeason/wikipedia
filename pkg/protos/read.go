package protos

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/willbeason/wikipedia/pkg/jobs"
	"google.golang.org/protobuf/proto"
)

// SizeLen is the number of bytes to use to store the length of the next proto
// message. At 4, is equivalent to uint32.
const SizeLen = 4

var ErrUnsupportedProtoExtension = errors.New("unsupported proto extension")

// ReadFile constructs a jobs.SourceFn from the passed file of protos.
func ReadFile[IN any, PIN Proto[IN]](path string) jobs.SourceFn[*IN] {
	return func(in chan<- *IN) jobs.Job {
		return readFileJob[IN, PIN](path, in)
	}
}

// ReadDir constructs a Source from the passed directory of files of protos.
func ReadDir[IN any, PIN Proto[IN]](dir string) jobs.SourceFn[*IN] {
	return func(in chan<- *IN) jobs.Job {
		return readDirJob[IN, PIN](dir, in)
	}
}

func readFileJob[IN any, PIN Proto[IN]](path string, in chan<- *IN) jobs.Job {
	return func(ctx context.Context, errs chan<- error) {
		readStream[IN, PIN](ctx, errs, path, in)
	}
}

func readDirJob[IN any, PIN Proto[IN]](dir string, in chan<- *IN) jobs.Job {
	return func(ctx context.Context, errs chan<- error) {
		files, err := os.ReadDir(dir)
		if err != nil {
			errs <- err
			return
		}

		for _, file := range files {
			path := filepath.Join(dir, file.Name())
			readStream[IN, PIN](ctx, errs, path, in)
		}
	}
}

func readStream[IN any, PIN Proto[IN]](ctx context.Context, errs chan<- error, path string, in chan<- *IN) {
	file, err := os.Open(path)
	if err != nil {
		errs <- err
		return
	}

	defer closeFile(file, errs)

	reader := bufio.NewReader(file)

ReadLoop:
	for {
		select {
		case <-ctx.Done():
			break ReadLoop
		default:
			out, _, readErr := readNextMessage[IN, PIN](reader)
			if readErr != nil {
				errs <- readErr
				break ReadLoop
			}

			in <- out
		}
	}
}

// readNextMessage reads the next Proto message from r.
// Returns the marshalled message and the number of bytes consumed from r.
func readNextMessage[IN any, PIN Proto[IN]](reader io.Reader) (*IN, int, error) {
	sizeBytes := make([]byte, SizeLen)
	if _, readErr := io.ReadFull(reader, sizeBytes); readErr != nil {
		if !errors.Is(readErr, io.EOF) {
			return nil, 0, fmt.Errorf("reading size of next message: %w", readErr)
		}
	}

	size := binary.LittleEndian.Uint32(sizeBytes)
	messageBytes := make([]byte, size)
	if _, readErr := io.ReadFull(reader, messageBytes); readErr != nil {
		return nil, 0, fmt.Errorf("reading next message bytes: %w", readErr)
	}

	var out PIN = new(IN)
	unmarshalErr := proto.Unmarshal(messageBytes, out)

	if unmarshalErr != nil {
		return nil, 0, fmt.Errorf("unmarshalling message: %w", unmarshalErr)
	}

	return out, SizeLen + int(size), nil
}
