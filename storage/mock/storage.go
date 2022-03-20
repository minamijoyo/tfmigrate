package mock

import (
	"context"
	"fmt"

	"github.com/minamijoyo/tfmigrate/storage"
)

// Storage is a storage.Storage implementation for mock.
// It writes and reads data from memory.
type Storage struct {
	// config is a storage config for mock
	config *Config
	// data stores a serialized data for history.
	data string
}

var _ storage.Storage = (*Storage)(nil)

// NewStorage returns a new instance of Storage.
func NewStorage(config *Config) (*Storage, error) {
	s := &Storage{
		config: config,
		data:   config.Data,
	}
	return s, nil
}

// Data returns a raw data in mock storage for testing.
func (s *Storage) Data() string {
	return s.data
}

// Write writes migration history data to storage.
func (s *Storage) Write(ctx context.Context, b []byte) error {
	if s.config.WriteError {
		return fmt.Errorf("failed to write mock storage: writeError = %t", s.config.WriteError)
	}
	s.data = string(b)
	return nil
}

// Read reads migration history data from storage.
func (s *Storage) Read(ctx context.Context) ([]byte, error) {
	if s.config.ReadError {
		return nil, fmt.Errorf("failed to read mock storage: readError = %t", s.config.ReadError)
	}
	return []byte(s.data), nil
}
