package history

import (
	"context"
	"io/ioutil"
)

// LocalStorage is an implementation of Storage for local file.
// This is only for debug and is not suitable for production use.
type LocalStorage struct {
	// Path to the migration history file. Relative to the current working directory.
	Path string
}

var _ Storage = (*LocalStorage)(nil)

// Write writes migration history data to storage.
func (s *LocalStorage) Write(ctx context.Context, b []byte) error {
	return ioutil.WriteFile(s.Path, b, 0644)
}

// Read reads migration history data from storage.
func (s *LocalStorage) Read(ctx context.Context) ([]byte, error) {
	return ioutil.ReadFile(s.Path)
}
