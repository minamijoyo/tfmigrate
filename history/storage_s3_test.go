package history

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

func TestS3StorageConfigNewStorage(t *testing.T) {
	cases := []struct {
		desc   string
		config *S3StorageConfig
		ok     bool
	}{
		{
			desc: "valid",
			config: &S3StorageConfig{
				Bucket: "tfmigrate-test",
				Key:    "tfmigrate/history.json",
			},
			ok: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := tc.config.NewStorage()
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error, got: %#v", got)
			}
			if tc.ok {
				_ = got.(*S3Storage)
			}
		})
	}
}

// mockS3Client is a mock implementation for testing.
type mockS3Client struct {
	putOutput *s3.PutObjectOutput
	getOutput *s3.GetObjectOutput
	err       error
}

// PutObjectWithContext returns a mocked response.
func (c *mockS3Client) PutObjectWithContext(ctx aws.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, error) {
	return c.putOutput, c.err
}

// GetObjectWithContext returns a mocked response.
func (c *mockS3Client) GetObjectWithContext(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error) {
	return c.getOutput, c.err
}

func TestS3StorageWrite(t *testing.T) {
	cases := []struct {
		desc     string
		option   *S3StorageOption
		bucket   string
		key      string
		contents []byte
		ok       bool
	}{
		{
			desc: "simple",
			option: &S3StorageOption{
				Client: &mockS3Client{
					putOutput: &s3.PutObjectOutput{},
					err:       nil,
				},
			},
			bucket:   "tfmigrate-test",
			key:      "tfmigrate/history.json",
			contents: []byte("foo"),
			ok:       true,
		},
		{
			desc: "bucket does not exist",
			option: &S3StorageOption{
				Client: &mockS3Client{
					putOutput: nil,
					err:       awserr.New("NoSuchBucket", "The specified bucket does not exist.", nil),
				},
			},
			bucket:   "not-exist-bucket",
			key:      "tfmigrate/history.json",
			contents: []byte("foo"),
			ok:       false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			s, err := NewS3Storage(tc.bucket, tc.key, tc.option)
			if err != nil {
				t.Fatalf("failed to NewS3Storage: %s", err)
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

func TestS3StorageRead(t *testing.T) {
	cases := []struct {
		desc     string
		option   *S3StorageOption
		bucket   string
		key      string
		contents []byte
		ok       bool
	}{
		{
			desc: "simple",
			option: &S3StorageOption{
				Client: &mockS3Client{
					getOutput: &s3.GetObjectOutput{
						Body: ioutil.NopCloser(strings.NewReader("foo")),
					},
					err: nil,
				},
			},
			bucket:   "tfmigrate-test",
			key:      "tfmigrate/history.json",
			contents: []byte("foo"),
			ok:       true,
		},
		{
			desc: "bucket does not exist",
			option: &S3StorageOption{
				Client: &mockS3Client{
					getOutput: nil,
					err:       awserr.New("NoSuchBucket", "The specified bucket does not exist.", nil),
				},
			},
			bucket:   "not-exist-bucket",
			key:      "tfmigrate/history.json",
			contents: nil,
			ok:       false,
		},
		{
			desc: "key does not exist",
			option: &S3StorageOption{
				Client: &mockS3Client{
					getOutput: nil,
					err:       awserr.New("NoSuchKey", "The specified key does not exist.", nil),
				},
			},
			bucket:   "tfmigrate-test",
			key:      "not_exist.json",
			contents: []byte{},
			ok:       true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			s, err := NewS3Storage(tc.bucket, tc.key, tc.option)
			if err != nil {
				t.Fatalf("failed to NewS3Storage: %s", err)
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
