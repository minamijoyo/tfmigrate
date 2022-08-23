package gcs

import (
	"context"
	"fmt"
	"io"
	"os"

	gcStorage "cloud.google.com/go/storage"
)

type Client interface {
	Read(ctx context.Context) ([]byte, error)
	Write(ctx context.Context, p []byte) error
}

type Adapter struct {
	config Config
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

func NewClient(ctx context.Context, config Config) (Client, error) {
	if config.Endpoint == "" {
		// https://pkg.go.dev/cloud.google.com/go/storage#hdr-Creating_a_Client
		err := os.Setenv("STORAGE_EMULATOR_HOST", config.Endpoint)
		if err != nil {
			return nil, err
		}
	}

	c, err := gcStorage.NewClient(ctx)
	a := &Adapter{
		config: config,
		client: c,
	}
	return a, err
}
