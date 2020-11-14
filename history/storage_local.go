package history

import (
	"context"
	"io/ioutil"
	"os"
)

// LocalStorageConfig is a config for local storage.
type LocalStorageConfig struct {
	// Path to a migration history file. Relative to the current working directory.
	Path string `hcl:"path"`
}

// LocalStorageConfig implements a StorageConfig.
var _ StorageConfig = (*LocalStorageConfig)(nil)

// NewStorage returns a new instance of LocalStorage.
func (c *LocalStorageConfig) NewStorage() (Storage, error) {
	s := NewLocalStorage(c.Path)
	return s, nil
}

// LocalStorage is an implementation of Storage for local file.
// This was originally intended for debugging purposes, but it can also be used
// as a workaround if Storage doesn't support your cloud provider.
// That is, you can manually synchronize local output files to the remote.
type LocalStorage struct {
	// path to a migration history file. Relative to the current working directory.
	path string
}

var _ Storage = (*LocalStorage)(nil)

// NewLocalStorage returns a new instance of LocalStorage.
func NewLocalStorage(path string) *LocalStorage {
	return &LocalStorage{
		path: path,
	}
}

// Write writes migration history data to storage.
func (s *LocalStorage) Write(ctx context.Context, b []byte) error {
	return ioutil.WriteFile(s.path, b, 0644)
}

// Read reads migration history data from storage.
// If the key does not exist, it is assumed to be uninitialized and returns
// an empty array instead of an error.
func (s *LocalStorage) Read(ctx context.Context) ([]byte, error) {
	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		// If the key does not exist
		return []byte{}, nil
	}
	return ioutil.ReadFile(s.path)
}
