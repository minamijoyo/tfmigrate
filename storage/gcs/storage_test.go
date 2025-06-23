package gcs

import (
	"context"
	"fmt"
	"testing"

	gcStorage "cloud.google.com/go/storage"
)

// mockClient is a mock implementation for testing.
type mockClient struct {
	dataToRead []byte
	err        error
	lockExists bool
	lockErr    error
	unlockErr  error
}

func (c *mockClient) Read(_ context.Context) ([]byte, error) {
	return c.dataToRead, c.err
}

func (c *mockClient) Write(_ context.Context, _ []byte) error {
	return c.err
}

func (c *mockClient) WriteLock(_ context.Context) error {
	if c.lockExists {
		return fmt.Errorf("lock file already exists")
	}
	if c.lockErr != nil {
		return c.lockErr
	}
	c.lockExists = true
	return nil
}

func (c *mockClient) Unlock(_ context.Context) error {
	if c.unlockErr != nil {
		return c.unlockErr
	}
	c.lockExists = false
	return nil
}

func TestStorageWrite(t *testing.T) {
	cases := []struct {
		desc     string
		config   *Config
		client   Client
		contents []byte
		ok       bool
	}{
		{
			desc: "simple",
			config: &Config{
				Bucket: "tfmigrate-test",
				Name:   "tfmigrate/history.json",
			},
			client: &mockClient{
				err: nil,
			},
			contents: []byte("foo"),
			ok:       true,
		},
		{
			desc: "bucket does not exist",
			config: &Config{
				Bucket: "not-exist-bucket",
				Name:   "tfmigrate/history.json",
			},
			client: &mockClient{
				err: gcStorage.ErrBucketNotExist,
			},
			contents: []byte("foo"),
			ok:       false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			s, err := NewStorage(tc.config, tc.client)
			if err != nil {
				t.Fatalf("failed to NewStorage: %s", err)
			}
			err = s.Write(context.Background(), tc.contents)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}
		})
	}
}

func TestStorageWriteLock(t *testing.T) {
	cases := []struct {
		desc       string
		config     *Config
		client     *mockClient
		ok         bool
		wantLocked bool
	}{
		{
			desc: "simple",
			config: &Config{
				Bucket: "tfmigrate-test",
				Name:   "tfmigrate/history.json",
			},
			client: &mockClient{
				lockExists: false,
				lockErr:    nil,
			},
			ok:         true,
			wantLocked: true,
		},
		{
			desc: "already locked",
			config: &Config{
				Bucket: "tfmigrate-test",
				Name:   "tfmigrate/history.json",
			},
			client: &mockClient{
				lockExists: true,
				lockErr:    nil,
			},
			ok:         false,
			wantLocked: true,
		},
		{
			desc: "error during lock",
			config: &Config{
				Bucket: "tfmigrate-test",
				Name:   "tfmigrate/history.json",
			},
			client: &mockClient{
				lockExists: false,
				lockErr:    fmt.Errorf("failed to create lock"),
			},
			ok:         false,
			wantLocked: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			s, err := NewStorage(tc.config, tc.client)
			if err != nil {
				t.Fatalf("failed to NewStorage: %s", err)
			}
			err = s.WriteLock(context.Background())
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}

			if tc.client.lockExists != tc.wantLocked {
				t.Errorf("got lock state: %v, want: %v", tc.client.lockExists, tc.wantLocked)
			}
		})
	}
}

func TestStorageUnlock(t *testing.T) {
	cases := []struct {
		desc       string
		config     *Config
		client     *mockClient
		ok         bool
		wantLocked bool
	}{
		{
			desc: "simple",
			config: &Config{
				Bucket: "tfmigrate-test",
				Name:   "tfmigrate/history.json",
			},
			client: &mockClient{
				lockExists: true,
				unlockErr:  nil,
			},
			ok:         true,
			wantLocked: false,
		},
		{
			desc: "not locked",
			config: &Config{
				Bucket: "tfmigrate-test",
				Name:   "tfmigrate/history.json",
			},
			client: &mockClient{
				lockExists: false,
				unlockErr:  nil,
			},
			ok:         true,
			wantLocked: false,
		},
		{
			desc: "error during unlock",
			config: &Config{
				Bucket: "tfmigrate-test",
				Name:   "tfmigrate/history.json",
			},
			client: &mockClient{
				lockExists: true,
				unlockErr:  fmt.Errorf("failed to unlock"),
			},
			ok:         false,
			wantLocked: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			s, err := NewStorage(tc.config, tc.client)
			if err != nil {
				t.Fatalf("failed to NewStorage: %s", err)
			}
			err = s.Unlock(context.Background())
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}

			if tc.client.lockExists != tc.wantLocked {
				t.Errorf("got lock state: %v, want: %v", tc.client.lockExists, tc.wantLocked)
			}
		})
	}
}

func TestStorageRead(t *testing.T) {
	cases := []struct {
		desc     string
		config   *Config
		client   Client
		contents []byte
		ok       bool
	}{
		{
			desc: "simple",
			config: &Config{
				Bucket: "tfmigrate-test",
				Name:   "tfmigrate/history.json",
			},
			client: &mockClient{
				dataToRead: []byte("foo"),
				err:        nil,
			},
			contents: []byte("foo"),
			ok:       true,
		},
		{
			desc: "bucket does not exist",
			config: &Config{
				Bucket: "not-exist-bucket",
				Name:   "tfmigrate/history.json",
			},
			client: &mockClient{
				dataToRead: nil,
				err:        gcStorage.ErrBucketNotExist,
			},
			contents: nil,
			ok:       false,
		},
		{
			desc: "object does not exist",
			config: &Config{
				Bucket: "tfmigrate-test",
				Name:   "not_exist.json",
			},
			client: &mockClient{
				err: gcStorage.ErrObjectNotExist,
			},
			contents: []byte{},
			ok:       true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			s, err := NewStorage(tc.config, tc.client)
			if err != nil {
				t.Fatalf("failed to NewStorage: %s", err)
			}
			got, err := s.Read(context.Background())
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}

			if tc.ok {
				if string(got) != string(tc.contents) {
					t.Errorf("got: %s, want: %s", string(got), string(tc.contents))
				}
			}
		})
	}
}
