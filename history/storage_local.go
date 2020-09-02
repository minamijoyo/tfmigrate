package history

import "io/ioutil"

// LocalStorage is an implementation of Storage for local file.
// This is only for debug and is not suitable for production use.
type LocalStorage struct {
	// path to the migration history file. Relative to the current working directory.
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
func (s *LocalStorage) Write(b []byte) error {
	return ioutil.WriteFile(s.path, b, 0644)
}

// Read reads migration history data from storage.
func (s *LocalStorage) Read() ([]byte, error) {
	return ioutil.ReadFile(s.path)
}
