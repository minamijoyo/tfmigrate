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
func (s *Storage) Write(_ context.Context, b []byte) error {
	if s.config.WriteError {
		return fmt.Errorf("failed to write mock storage: writeError = %t", s.config.WriteError)
	}
	s.data = string(b)
	return nil
}

// WriteLock writes a lock file to storage.
func (s *Storage) WriteLock(_ context.Context) error {
	if s.config.WriteError {
		return fmt.Errorf("failed to write lock in mock storage: writeError = %t", s.config.WriteError)
	}

	// Check if lock exists in config
	if s.config.LockExists {
		return fmt.Errorf("lock file already exists in mock storage")
	}

	// Set lock exists
	s.config.LockExists = true
	return nil
}

// Unlock removes a lock file from storage.
func (s *Storage) Unlock(_ context.Context) error {
	if s.config.WriteError {
		return fmt.Errorf("failed to remove lock from mock storage: writeError = %t", s.config.WriteError)
	}

	// Reset lock status
	s.config.LockExists = false
	return nil
}

// Read reads migration history data from storage.
func (s *Storage) Read(_ context.Context) ([]byte, error) {
	if s.config.ReadError {
		return nil, fmt.Errorf("failed to read mock storage: readError = %t", s.config.ReadError)
	}
	return []byte(s.data), nil
}
