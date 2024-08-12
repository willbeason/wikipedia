package flags

import (
	"errors"
	"fmt"
	"os"
)

var ErrDirectory = errors.New("unable to use directory")

func CreateOrCheckDirectory(path string) error {
	if entries, err := os.ReadDir(path); !os.IsNotExist(err) {
		if err != nil {
			return fmt.Errorf("%w: unable to determine if directory already exists at %q: %w",
				ErrDirectory, path, err)
		}

		if len(entries) > 0 {
			return fmt.Errorf("%w: out directory exists and is non-empty: %q",
				ErrDirectory, path)
		}
	} else {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return fmt.Errorf("creating output directory %q: %w", path, err)
		}
	}

	return nil
}
