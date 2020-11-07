package history

import (
	"context"
	"fmt"
)

// MockStorageConfig is a config for mock storage.
type MockStorageConfig struct {
	// Data stores a serialized data for history.
	Data string `hcl:"data"`
	// WriteError is a flag to return an error on Write().
	WriteError bool `hcl:"write_error"`
	// ReadError is a flag to return an error on Read().
	ReadError bool `hcl:"read_error"`
}

// MockStorageConfig implements a StorageConfig.
var _ StorageConfig = (*MockStorageConfig)(nil)

// NewStorage returns a new instance of MockStorage.
func (c *MockStorageConfig) NewStorage() (Storage, error) {
	s := NewMockStorage(c.Data, c.WriteError, c.ReadError)
	return s, nil
}

// MockStorage is an implementation of Storage for testing.
// It writes and reads data from memory.
type MockStorage struct {
	// data stores a serialized data for history.
	data string
	// writeError is a flag to return an error on Write().
	writeError bool
	// readError is a flag to return an error on Read().
	readError bool
}

var _ Storage = (*MockStorage)(nil)

// NewMockStorage returns a new instance of MockStorage.
func NewMockStorage(data string, writeError bool, readError bool) *MockStorage {
	return &MockStorage{
		data:       data,
		writeError: writeError,
		readError:  readError,
	}
}

// Write writes migration history data to storage.
func (s *MockStorage) Write(ctx context.Context, b []byte) error {
	if s.writeError {
		return fmt.Errorf("failed to write mock storage: werr = %t", s.writeError)
	}
	s.data = string(b)
	return nil
}

// Read reads migration history data from storage.
func (s *MockStorage) Read(ctx context.Context) ([]byte, error) {
	if s.readError {
		return nil, fmt.Errorf("failed to read mock storage: rerr = %t", s.readError)
	}
	return []byte(s.data), nil
}
