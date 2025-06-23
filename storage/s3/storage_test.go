package s3

import (
	// mockClient is a mock implementation for testing.
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// mockClient is a mock implementation for testing.
type mockClient struct {
	putOutput    *s3.PutObjectOutput
	getOutput    *s3.GetObjectOutput
	deleteOutput *s3.DeleteObjectOutput
	err          error
}

// PutObjectWithContext returns a mocked response.
func (c *mockClient) PutObject(_ context.Context, _ *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return c.putOutput, c.err
}

// GetObjectWithContext returns a mocked response.
func (c *mockClient) GetObject(_ context.Context, _ *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return c.getOutput, c.err
}

// DeleteObject returns a mocked response for DeleteObject.
func (c *mockClient) DeleteObject(_ context.Context, _ *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	return &s3.DeleteObjectOutput{}, c.err
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
				Key:    "tfmigrate/history.json",
			},
			client: &mockClient{
				putOutput: &s3.PutObjectOutput{},
				err:       nil,
			},
			contents: []byte("foo"),
			ok:       true,
		},
		{
			desc: "bucket does not exist",
			config: &Config{
				Bucket: "not-exist-bucket",
				Key:    "tfmigrate/history.json",
			},
			client: &mockClient{
				putOutput: nil,
				err:       &types.NoSuchBucket{Message: aws.String("The specified bucket does not exist.")},
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
		desc   string
		config *Config
		client Client
		ok     bool
	}{
		{
			desc: "simple",
			config: &Config{
				Bucket: "tfmigrate-test",
				Key:    "tfmigrate/history.json",
			},
			client: &mockClient{
				putOutput: &s3.PutObjectOutput{},
				err:       nil,
			},
			ok: true,
		},
		{
			desc: "lock already exists",
			config: &Config{
				Bucket: "tfmigrate-test",
				Key:    "tfmigrate/history.json",
			},
			client: &mockClient{
				putOutput: nil,
				err:       fmt.Errorf("PreconditionFailed: The condition specified in the request evaluated to false"),
			},
			ok: false,
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
		})
	}
}

func TestStorageUnlock(t *testing.T) {
	cases := []struct {
		desc   string
		config *Config
		client Client
		ok     bool
	}{
		{
			desc: "simple",
			config: &Config{
				Bucket: "tfmigrate-test",
				Key:    "tfmigrate/history.json",
			},
			client: &mockClient{
				err: nil,
			},
			ok: true,
		},
		{
			desc: "lock does not exist",
			config: &Config{
				Bucket: "tfmigrate-test",
				Key:    "tfmigrate/history.json",
			},
			client: &mockClient{
				err: &types.NoSuchKey{},
			},
			ok: true,
		},
		{
			desc: "other error",
			config: &Config{
				Bucket: "tfmigrate-test",
				Key:    "tfmigrate/history.json",
			},
			client: &mockClient{
				err: fmt.Errorf("some other error"),
			},
			ok: false,
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
				Key:    "tfmigrate/history.json",
			},
			client: &mockClient{
				getOutput: &s3.GetObjectOutput{
					Body: io.NopCloser(strings.NewReader("foo")),
				},
				err: nil,
			},
			contents: []byte("foo"),
			ok:       true,
		},
		{
			desc: "bucket does not exist",
			config: &Config{
				Bucket: "not-exist-bucket",
				Key:    "tfmigrate/history.json",
			},
			client: &mockClient{
				getOutput: nil,
				err:       &types.NoSuchBucket{Message: aws.String("The specified bucket does not exist.")},
			},
			contents: nil,
			ok:       false,
		},
		{
			desc: "key does not exist",
			config: &Config{
				Bucket: "tfmigrate-test",
				Key:    "not_exist.json",
			},
			client: &mockClient{
				getOutput: nil,
				err:       &types.NoSuchKey{Message: aws.String("The specified key does not exist.")},
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
