package mock

import (
	"context"
	"fmt"

	"github.com/minamijoyo/tfmigrate/storage"
)

// Storage is a storage.Storage implementation for testing.
// It writes and reads data from memory.
type Storage struct {
	// data stores a serialized data for history.
	data string
	// writeError is a flag to return an error on Write().
	writeError bool
	// readError is a flag to return an error on Read().
	readError bool
}

var _ storage.Storage = (*Storage)(nil)

// NewStorage returns a new instance of Storage.
func NewStorage(data string, writeError bool, readError bool) *Storage {
	return &Storage{
		data:       data,
		writeError: writeError,
		readError:  readError,
	}
}

// Write writes migration history data to storage.
func (s *Storage) Write(ctx context.Context, b []byte) error {
	if s.writeError {
		return fmt.Errorf("failed to write mock storage: writeError = %t", s.writeError)
	}
	s.data = string(b)
	return nil
}

// Read reads migration history data from storage.
func (s *Storage) Read(ctx context.Context) ([]byte, error) {
	if s.readError {
		return nil, fmt.Errorf("failed to read mock storage: readError = %t", s.readError)
	}
	return []byte(s.data), nil
}
