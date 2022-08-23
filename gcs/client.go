package gcs

import (
	"context"
	"fmt"
	"io"

	gcStorage "cloud.google.com/go/storage"
)

// A minimal interface to mock behavior of GCS client.
type Client interface {
	// Read an object from a GCS bucket.
	Read(ctx context.Context) ([]byte, error)

	// Write an object onto a GCS bucket.
	Write(ctx context.Context, p []byte) error
}

// An implementation of Client that delegates actual operation to gcsStorage.Client.
type Adapter struct {
	// A config to specify which bucket and object we handle.
	config Config
	// A GCS client which is delegated actual operation.
	client *gcStorage.Client
}

func (a Adapter) Read(ctx context.Context) ([]byte, error) {
	r, err := a.client.Bucket(a.config.Bucket).Object(a.config.Name).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	body, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed reading from gcs://%s/%s: %w", a.config.Bucket, a.config.Name, err)
	}
	return body, nil
}

func (a Adapter) Write(ctx context.Context, p []byte) error {
	w := a.client.Bucket(a.config.Bucket).Object(a.config.Name).NewWriter(ctx)
	_, err := w.Write(p)

	if err != nil {
		return fmt.Errorf("failed writing to gcs://%s/%s: %w", a.config.Bucket, a.config.Name, err)
	}
	return w.Close()
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
