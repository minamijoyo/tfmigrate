package gcs

import (
	"context"
	"fmt"
	"io"

	gcStorage "cloud.google.com/go/storage"
)

// Client is an interface for GCS storage client.
type Client interface {
	Write(ctx context.Context, b []byte) error
	Read(ctx context.Context) ([]byte, error)
	WriteLock(ctx context.Context) error
	Unlock(ctx context.Context) error
}

// Adapter implements the Client interface with *storage.Client.
type Adapter struct {
	config Config
	client *gcStorage.Client
}

var _ Client = (*Adapter)(nil)

// Write writes a blob to GCS.
func (a Adapter) Write(ctx context.Context, b []byte) error {
	bkt := a.client.Bucket(a.config.Bucket)
	obj := bkt.Object(a.config.Name)
	w := obj.NewWriter(ctx)
	if _, err := w.Write(b); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return nil
}

// Read reads a blob from GCS.
func (a Adapter) Read(ctx context.Context) ([]byte, error) {
	bkt := a.client.Bucket(a.config.Bucket)
	obj := bkt.Object(a.config.Name)

	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// WriteLock creates a lock file in the GCS bucket.
func (a Adapter) WriteLock(ctx context.Context) error {
	bkt := a.client.Bucket(a.config.Bucket)
	lockObjName := a.config.Name + ".lock"
	obj := bkt.Object(lockObjName)

	// Check if lock file already exists
	_, err := obj.Attrs(ctx)
	if err == nil {
		// Lock file exists
		return fmt.Errorf("lock file already exists, another process may be running")
	}

	if err != gcStorage.ErrObjectNotExist {
		// Some other error occurred
		return fmt.Errorf("failed to check lock file: %w", err)
	}

	// Create lock file
	w := obj.NewWriter(ctx)
	if _, err := w.Write([]byte("locked")); err != nil {
		return fmt.Errorf("failed to write lock file: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close lock file writer: %w", err)
	}

	return nil
}

// Unlock removes the lock file from the GCS bucket.
func (a Adapter) Unlock(ctx context.Context) error {
	bkt := a.client.Bucket(a.config.Bucket)
	lockObjName := a.config.Name + ".lock"
	obj := bkt.Object(lockObjName)

	// Delete the lock file
	err := obj.Delete(ctx)
	if err != nil && err != gcStorage.ErrObjectNotExist {
		// Error if delete fails, but ignore if the file doesn't exist
		return fmt.Errorf("failed to delete lock file: %w", err)
	}

	return nil
}

// NewClient returns a new Client with given Context and Config.
func NewClient(ctx context.Context, config Config) (Client, error) {
	c, err := gcStorage.NewClient(ctx)
	a := &Adapter{
		config: config,
		client: c,
	}
	return a, err
}
