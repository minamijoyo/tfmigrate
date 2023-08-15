package gcs

import (
	"context"
	"testing"

	gcStorage "cloud.google.com/go/storage"
)

// mockClient is a mock implementation for testing.
type mockClient struct {
	dataToRead []byte
	err        error
}

func (c *mockClient) Read(ctx context.Context) ([]byte, error) {
	return c.dataToRead, c.err
}

func (c *mockClient) Write(ctx context.Context, b []byte) error {
	return c.err
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
