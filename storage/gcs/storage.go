package gcs

import (
	"context"

	gcStorage "cloud.google.com/go/storage"
	"github.com/minamijoyo/tfmigrate/storage"
)

// An implementation of [storage.Storage] interface.
type Storage struct {
	// config is a storage config for GCS.
	config *Config
	// client is an instance of Client interface to call API.
	// It is intended to be replaced with a mock for testing.
	// https://pkg.go.dev/cloud.google.com/go/storage#Client
	client Client
}

// WriteLock implements storage.Storage.
// It creates a lock file in the GCS bucket to prevent concurrent modifications.
func (s *Storage) WriteLock(ctx context.Context) error {
	err := s.init(ctx)
	if err != nil {
		return err
	}

	return s.client.WriteLock(ctx)
}

// Unlock implements storage.Storage.
// It removes the lock file from the GCS bucket.
func (s *Storage) Unlock(ctx context.Context) error {
	err := s.init(ctx)
	if err != nil {
		return err
	}

	return s.client.Unlock(ctx)
}

var _ storage.Storage = (*Storage)(nil)

// NewStorage returns a new instance of Storage.
func NewStorage(config *Config, client Client) (*Storage, error) {
	s := &Storage{
		config: config,
		client: client,
	}
	return s, nil
}

func (s *Storage) Write(ctx context.Context, b []byte) error {
	err := s.init(ctx)
	if err != nil {
		return err
	}

	return s.client.Write(ctx, b)
}

func (s *Storage) Read(ctx context.Context) ([]byte, error) {
	err := s.init(ctx)
	if err != nil {
		return nil, err
	}

	r, err := s.client.Read(ctx)
	if err == gcStorage.ErrObjectNotExist {
		return []byte{}, nil
	} else if err != nil {
		return nil, err
	}
	return r, nil
}

func (s *Storage) init(ctx context.Context) error {
	if s.client == nil {
		client, err := gcStorage.NewClient(ctx)
		if err != nil {
			return err
		}
		s.client = Adapter{
			config: *s.config,
			client: client,
		}
	}
	return nil
}
