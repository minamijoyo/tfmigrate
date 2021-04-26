package history

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/googleapis/google-cloud-go-testing/storage/stiface"
	"google.golang.org/api/googleapi"
)

func TestGCSStorageConfigNewStorage(t *testing.T) {
	cases := []struct {
		desc   string
		config *GCSStorageConfig
		ok     bool
	}{
		{
			desc: "valid",
			config: &GCSStorageConfig{
				Bucket:                    "tfmigrate-test",
				Prefix:                    "tfmigrate/history.json",
				Credentials:               "dummy",
				AccessToken:               "dummy",
				ImpersonateServiceAccount: "dev",
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
				_ = got.(*GCSStorage)
			}
		})
	}
}

// mockGCSClient is a mock implementation for testing.
type mockGCSClient struct {
	stiface.Client
	buckets map[string]map[string]*bytes.Buffer
	err     error
}

func (c mockGCSClient) Bucket(name string) stiface.BucketHandle {
	return mockGCSBucket{
		client:     c,
		bucketName: name,
	}
}

func mockGCSNotFound(bucketName, objectName string) error {
	msg := "Not Found"
	if objectName != "" {
		return errors.New("storage: object doesn't exist")
	}

	return &googleapi.Error{
		Code:    http.StatusNotFound,
		Message: msg,
		Errors: []googleapi.ErrorItem{
			{
				Reason:  "notFound",
				Message: msg,
			},
		},
	}
}

// mockGCSBucket allows access to buckets by name.
type mockGCSBucket struct {
	stiface.BucketHandle
	client     mockGCSClient
	bucketName string
}

func (b mockGCSBucket) Object(name string) stiface.ObjectHandle {
	return mockGCSObject{
		client:     b.client,
		bucketName: b.bucketName,
		objectName: name,
	}
}

// mockGCSObject allows access to objects by name.
type mockGCSObject struct {
	stiface.ObjectHandle
	client     mockGCSClient
	bucketName string
	objectName string
}

func (o mockGCSObject) NewReader(ctx context.Context) (stiface.Reader, error) {
	if bucket, ok := o.client.buckets[o.bucketName]; !ok {
		return nil, mockGCSNotFound(o.bucketName, "")
	} else if objectData, ok := bucket[o.objectName]; !ok || objectData == nil {
		return nil, mockGCSNotFound(o.bucketName, o.objectName)
	} else {
		return mockGCSReader{
			buffer: objectData,
		}, nil
	}
}
func (o mockGCSObject) NewWriter(ctx context.Context) stiface.Writer {
	bucket, ok := o.client.buckets[o.bucketName]
	if !ok {
		return mockGCSWriter{
			bucketName: o.bucketName,
		}
	}

	bucket[o.objectName] = &bytes.Buffer{}
	return mockGCSWriter{
		buffer: bucket[o.objectName],
	}
}

type mockGCSReader struct {
	stiface.Reader
	buffer *bytes.Buffer
}

func (r mockGCSReader) Read(p []byte) (n int, err error) {
	return r.buffer.Read(p)
}
func (r mockGCSReader) Close() error {
	return nil
}

type mockGCSWriter struct {
	stiface.Writer
	buffer     *bytes.Buffer
	bucketName string
}

func (w mockGCSWriter) Write(p []byte) (n int, err error) {
	if w.buffer == nil {
		return 0, mockGCSNotFound(w.bucketName, "")
	}
	return w.buffer.Write(p)
}
func (w mockGCSWriter) Close() error {
	return nil
}

func TestGCSStorageWrite(t *testing.T) {
	cases := []struct {
		desc     string
		config   *GCSStorageConfig
		client   stiface.Client
		contents []byte
		ok       bool
	}{
		{
			desc: "simple",
			config: &GCSStorageConfig{
				Bucket: "tfmigrate-test",
				Prefix: "tfmigrate/history.json",
			},
			client: &mockGCSClient{
				buckets: map[string]map[string]*bytes.Buffer{
					"tfmigrate-test": {},
				},
				err: nil,
			},
			contents: []byte("foo"),
			ok:       true,
		},
		{
			desc: "bucket does not exist",
			config: &GCSStorageConfig{
				Bucket: "not-exist-bucket",
				Prefix: "tfmigrate/history.json",
			},
			client: &mockGCSClient{
				err: mockGCSNotFound("not-exist-bucket", ""),
			},
			contents: []byte("foo"),
			ok:       false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			s, err := NewGCSStorage(tc.config, tc.client)
			if err != nil {
				t.Fatalf("failed to NewGCSStorage: %s", err)
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

func TestGCSStorageRead(t *testing.T) {
	cases := []struct {
		desc     string
		config   *GCSStorageConfig
		client   stiface.Client
		contents []byte
		ok       bool
	}{
		{
			desc: "simple",
			config: &GCSStorageConfig{
				Bucket: "tfmigrate-test",
				Prefix: "tfmigrate/history.json",
			},
			client: &mockGCSClient{
				buckets: map[string]map[string]*bytes.Buffer{
					"tfmigrate-test": {
						"tfmigrate/history.json": bytes.NewBufferString("foo"),
					},
				},
				err: nil,
			},
			contents: []byte("foo"),
			ok:       true,
		},
		{
			desc: "bucket does not exist",
			config: &GCSStorageConfig{
				Bucket: "not-exist-bucket",
				Prefix: "tfmigrate/history.json",
			},
			client: &mockGCSClient{
				err: mockGCSNotFound("not-exist-bucket", ""),
			},
			contents: nil,
			ok:       false,
		},
		{
			desc: "key does not exist",
			config: &GCSStorageConfig{
				Bucket: "tfmigrate-test",
				Prefix: "not_exist.json",
			},
			client: &mockGCSClient{
				buckets: map[string]map[string]*bytes.Buffer{
					"tfmigrate-test": {},
				},
				err: mockGCSNotFound("not-exist-bucket", "not_exist.json"),
			},
			contents: []byte{},
			ok:       true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			s, err := NewGCSStorage(tc.config, tc.client)
			if err != nil {
				t.Fatalf("failed to NewGCSStorage: %s", err)
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
