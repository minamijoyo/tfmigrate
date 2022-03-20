package s3

import (
	"context"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
)

// mockClient is a mock implementation for testing.
type mockClient struct {
	putOutput *s3.PutObjectOutput
	getOutput *s3.GetObjectOutput
	err       error
}

// PutObjectWithContext returns a mocked response.
func (c *mockClient) PutObjectWithContext(ctx aws.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, error) {
	return c.putOutput, c.err
}

// GetObjectWithContext returns a mocked response.
func (c *mockClient) GetObjectWithContext(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error) {
	return c.getOutput, c.err
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
				err:       awserr.New("NoSuchBucket", "The specified bucket does not exist.", nil),
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
				Key:    "tfmigrate/history.json",
			},
			client: &mockClient{
				getOutput: &s3.GetObjectOutput{
					Body: ioutil.NopCloser(strings.NewReader("foo")),
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
				err:       awserr.New("NoSuchBucket", "The specified bucket does not exist.", nil),
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
				err:       awserr.New("NoSuchKey", "The specified key does not exist.", nil),
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
