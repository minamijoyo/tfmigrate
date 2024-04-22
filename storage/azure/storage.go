package azure

import (
	"context"

	"github.com/minamijoyo/tfmigrate/storage"
)

// Storage is a storage.Storage implementation for Azure Blob Storage.
type Storage struct {
	// config is a storage config for azure blob storage.
	config *Config
	// client is an instance of azclient interface to call API.
	// It is intended to be replaced with a mock for testing.
	client Client
}

var _ storage.Storage = (*Storage)(nil)

// NewStorage returns a new instance of Storage.
func NewStorage(config *Config, client Client) (*Storage, error) {
	if client == nil {
		var err error
		client, err = newClient(config)
		if err != nil {
			return nil, err
		}
	}

	s := &Storage{
		config: config,
		client: client,
	}

	return s, nil
}

// Write writes migration history data to storage.
func (s *Storage) Write(ctx context.Context, b []byte) error {
	return s.client.Write(ctx, s.config.ContainerName, s.config.Key, b)
}

// Read reads migration history data from storage.
func (s *Storage) Read(ctx context.Context) ([]byte, error) {
	return s.client.Read(ctx, s.config.ContainerName, s.config.Key)
}
