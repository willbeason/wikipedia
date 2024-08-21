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

	progress_bar "github.com/willbeason/wikipedia/pkg/progress-bar"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

// SizeLen is the number of bytes to use to store the length of the next proto
// message. At 4, is equivalent to uint32.
const SizeLen = 4

var ErrUnsupportedProtoExtension = errors.New("unsupported proto extension")

func ReadStream[OUT any, POUT Proto[OUT]](
	ctx context.Context,
	cancel context.CancelCauseFunc,
	file string,
) <-chan *OUT {
	result := make(chan *OUT, 100)

	go func() {
		f, err := os.Open(file)
		defer func() {
			closeErr := f.Close()
			if closeErr != nil {
				fmt.Printf("closing %q: %v\n", file, closeErr)
			}
		}()

		if err != nil {
			cancel(err)
			return
		}

		fInfo, err := f.Stat()
		if err != nil {
			cancel(err)
			return
		}

		fileSize := fInfo.Size()
		fileProgress := int64(0)
		fileProgressBar := progress_bar.NewProgressBar("Loading "+fInfo.Name(), fileSize, os.Stdout)
		fileProgressBar.Start()
		defer func() {
			fileProgressBar.Stop()
		}()

		bf := bufio.NewReader(f)

	ScanLoop:
		for {
			select {
			case <-ctx.Done():
				break ScanLoop
			default:
				sizeBytes := make([]byte, SizeLen)
				if _, readErr := io.ReadFull(bf, sizeBytes); readErr != nil {
					if !errors.Is(readErr, io.EOF) {
						cancel(fmt.Errorf("reading size of next message: %w", readErr))
					}
					break ScanLoop
				}

				size := binary.LittleEndian.Uint32(sizeBytes)

				messageBytes := make([]byte, size)
				if _, readErr := io.ReadFull(bf, messageBytes); readErr != nil {
					cancel(fmt.Errorf("reading next message: %w", readErr))
					break ScanLoop
				}

				fileProgress += int64(len(messageBytes)) + SizeLen
				fileProgressBar.Set(fileProgress)

				var out POUT = new(OUT)
				unmarshalErr := proto.Unmarshal(messageBytes, out)

				if unmarshalErr != nil {
					cancel(fmt.Errorf("unmarshalling message: %w", unmarshalErr))
					break ScanLoop
				}

				result <- out
			}
		}

		close(result)
	}()

	return result
}

// Read reads a protocol buffer stored in file to a protocol buffer of type OUT.
// OUT must be a type whose pointer receiver is a proto.Message.
func Read[OUT any, POUT Proto[OUT]](file string) (*OUT, error) {
	var out POUT = new(OUT)

	bytes, err := os.ReadFile(file)
	if err != nil {
		return out, fmt.Errorf("reading %q: %w", file, err)
	}

	switch ext := filepath.Ext(file); ext {
	case ".pb":
		err = proto.Unmarshal(bytes, out)
	case ".json":
		err = protojson.Unmarshal(bytes, out)
	case ".txt":
		err = prototext.Unmarshal(bytes, out)
	default:
		return out, fmt.Errorf("%w: %q", ErrUnsupportedProtoExtension, ext)
	}

	if err != nil {
		return out, fmt.Errorf("unmarshalling %q: %w", file, err)
	}

	return out, nil
}

func Write(path string, p proto.Message) error {
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return fmt.Errorf("writing %q: %w", path, err)
	}

	var bytes []byte

	switch ext := filepath.Ext(path); ext {
	case ".pb":
		bytes, err = proto.Marshal(p)
	case ".json":
		bytes, err = protojson.MarshalOptions{Indent: "  "}.Marshal(p)
	case ".txt":
		bytes, err = prototext.MarshalOptions{Indent: "  "}.Marshal(p)
	default:
		return fmt.Errorf("%w: %q", ErrUnsupportedProtoExtension, ext)
	}
	if err != nil {
		return fmt.Errorf("marshalling %q: %w", path, err)
	}

	err = os.WriteFile(path, bytes, os.ModePerm)
	if err != nil {
		return fmt.Errorf("writing %q: %w", path, err)
	}

	return nil
}
