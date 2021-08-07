package walker

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"time"
)

// Files returns a function which writes file paths to work.
func Files(work chan<- string) func(string, fs.DirEntry, error) error {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			fmt.Println(time.Now().Format(time.RFC3339), path)
			return nil
		}

		work <- filepath.ToSlash(path)

		return nil
	}
}
