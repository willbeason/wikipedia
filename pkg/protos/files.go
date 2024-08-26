package protos

import (
	"fmt"
	"os"
)

func closeFile(file *os.File, errs chan<- error) {
	err := file.Close()
	if err != nil {
		errs <- fmt.Errorf("closing %q: %w", file.Name(), err)
	}
}
