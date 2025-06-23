package local

import (
	"context"
	"fmt"
	"os"

	"github.com/minamijoyo/tfmigrate/storage"
)

// Storage is a storage.Storage implementation for local file.
// This was originally intended for debugging purposes, but it can also be used
// as a workaround if Storage doesn't support your cloud provider.
// That is, you can manually synchronize local output files to the remote.
type Storage struct {
	// config is a storage config for local.
	config *Config
}

var _ storage.Storage = (*Storage)(nil)

// NewStorage returns a new instance of Storage.
func NewStorage(config *Config) (*Storage, error) {
	s := &Storage{
		config: config,
	}
	return s, nil
}

// Write writes migration history data to storage.
func (s *Storage) Write(_ context.Context, b []byte) error {
	// nolint gosec
	// G306: Expect WriteFile permissions to be 0600 or less
	// We ignore it because a history file doesn't contains sensitive data.
	// Note that changing a permission to 0600 is breaking change.
	return os.WriteFile(s.config.Path, b, 0644)
}

// WriteLock writes a lock file to local storage.
func (s *Storage) WriteLock(_ context.Context) error {
	lockPath := s.config.Path + ".lock"
	// Check if lock file already exists
	if _, err := os.Stat(lockPath); err == nil {
		// Lock file already exists
		return fmt.Errorf("lock file already exists at %s", lockPath)
	} else if !os.IsNotExist(err) {
		// An unexpected error occurred
		return fmt.Errorf("failed to check lock file at %s: %w", lockPath, err)
	}

	// Create lock file
	f, err := os.Create(lockPath)
	if err != nil {
		return fmt.Errorf("failed to create lock file at %s: %w", lockPath, err)
	}
	return f.Close()
}

// Unlock removes a lock file from local storage.
func (s *Storage) Unlock(_ context.Context) error {
	lockPath := s.config.Path + ".lock"
	// If lock file doesn't exist, it's not an error
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		return nil
	}
	err := os.Remove(lockPath)
	if err != nil {
		return fmt.Errorf("failed to remove lock file at %s: %w", lockPath, err)
	}
	return nil
}

// Read reads migration history data from storage.
// If the key does not exist, it is assumed to be uninitialized and returns
// an empty array instead of an error.
func (s *Storage) Read(_ context.Context) ([]byte, error) {
	if _, err := os.Stat(s.config.Path); os.IsNotExist(err) {
		// If the key does not exist
		return []byte{}, nil
	}
	return os.ReadFile(s.config.Path)
}
