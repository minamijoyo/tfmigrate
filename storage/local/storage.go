package local

import (
	"context"
	"io/ioutil"
	"os"

	storage "github.com/minamijoyo/tfmigrate-storage"
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
func (s *Storage) Write(ctx context.Context, b []byte) error {
	// nolint gosec
	// G306: Expect WriteFile permissions to be 0600 or less
	// We ignore it because a history file doesn't contains sensitive data.
	// Note that changing a permission to 0600 is breaking change.
	return ioutil.WriteFile(s.config.Path, b, 0644)
}

// Read reads migration history data from storage.
// If the key does not exist, it is assumed to be uninitialized and returns
// an empty array instead of an error.
func (s *Storage) Read(ctx context.Context) ([]byte, error) {
	if _, err := os.Stat(s.config.Path); os.IsNotExist(err) {
		// If the key does not exist
		return []byte{}, nil
	}
	return ioutil.ReadFile(s.config.Path)
}
