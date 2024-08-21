package protos

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"sync"

	"google.golang.org/protobuf/proto"
)

func WriteStream[IN any, PIN Proto[IN]](
	ctx context.Context,
	cancel context.CancelCauseFunc,
	filepath string,
	protos <-chan PIN,
) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		f, err := os.Create(filepath)
		if err != nil {
			cancel(err)
			return
		}

		defer func() {
			closeErr := f.Close()
			if closeErr != nil {
				fmt.Printf("closing %q: %v\n", filepath, closeErr)
			}
			wg.Done()
		}()

		bf := bufio.NewWriter(f)

	WriteLoop:
		for {
			select {
			case <-ctx.Done():
				break WriteLoop
			case p, ok := <-protos:
				if !ok {
					// Protos channel is closed.
					flushErr := bf.Flush()
					if flushErr != nil {
						fmt.Printf("flushing %q: %v\n", filepath, flushErr)
					}
					break WriteLoop
				}

				bytes, marshalErr := proto.Marshal(p)
				if marshalErr != nil {
					cancel(marshalErr)
					break WriteLoop
				}

				// Record length of marshaled message.
				lengthBuf := make([]byte, SizeLen)
				binary.LittleEndian.PutUint32(lengthBuf, uint32(len(bytes)))
				if _, writeErr := bf.Write(lengthBuf); writeErr != nil {
					cancel(writeErr)
					break WriteLoop
				}

				// Record message.
				if _, writeErr := bf.Write(bytes); writeErr != nil {
					cancel(writeErr)
					break WriteLoop
				}
			}
		}
	}()

	return wg
}
