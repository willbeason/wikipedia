package db

import (
	"encoding/binary"
)

func NewRunner(path string, parallel int) *Runner {
	return &Runner{path: path, parallel: parallel}
}

// Runner provides streaming operations on badger.DB.
type Runner struct {
	// parallel is the maximum number of goroutines for an operation.
	parallel int
	// path is the path to a badger database.
	path string
}

func toKey(id uint32) []byte {
	const uint32Bytes = 4

	key := make([]byte, uint32Bytes)
	binary.LittleEndian.PutUint32(key, id)

	return key
}
