package s3

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/minamijoyo/tfmigrate/storage"

	awsbase "github.com/hashicorp/aws-sdk-go-base"
)

// Config is a config for s3 storage.
// This is expected to have almost the same options as Terraform s3 backend.
// https://www.terraform.io/docs/backends/types/s3.html
// However, it has many minor options and it's a pain to test all options from
// first, so we added only options we need for now.
type Config struct {
	// Name of the bucket.
	Bucket string `hcl:"bucket"`
	// Path to the migration history file.
	Key string `hcl:"key"`

	// AWS region.
	Region string `hcl:"region,optional"`
	// Custom endpoint for the AWS S3 API.
	Endpoint string `hcl:"endpoint,optional"`
	// AWS access key.
	AccessKey string `hcl:"access_key,optional"`
	// AWS secret key.
	SecretKey string `hcl:"secret_key,optional"`
	// Name of AWS profile in AWS shared credentials file.
	Profile string `hcl:"profile,optional"`
	// Amazon Resource Name (ARN) of the IAM Role to assume.
	RoleARN string `hcl:"role_arn,optional"`
	// Skip credentials validation via the STS API.
	SkipCredentialsValidation bool `hcl:"skip_credentials_validation,optional"`
	// Skip usage of EC2 Metadata API.
	SkipMetadataAPICheck bool `hcl:"skip_metadata_api_check,optional"`
	// Enable path-style S3 URLs (https://<HOST>/<BUCKET>
	// instead of https://<BUCKET>.<HOST>).
	ForcePathStyle bool `hcl:"force_path_style,optional"`
	// SSE KMS Key Id for optional server-side encryption enablement
	KmsKeyID string `hcl:"kms_key_id,optional"`
}

// Config implements a storage.Config.
var _ storage.Config = (*Config)(nil)

// NewStorage returns a new instance of storage.Storage.
func (c *Config) NewStorage() (storage.Storage, error) {
	return NewStorage(c, nil)
}

// Client is an abstraction layer for AWS S3 API.
// It is intended to be replaced with a mock for testing.
type Client interface {
	// PutObjectWithContext puts a file to S3.
	PutObjectWithContext(ctx aws.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, error)
	// GetObjectWithContext gets a file from S3.
	GetObjectWithContext(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error)
}

// client is a real implementation of the Client.
type client struct {
	s3api s3iface.S3API
}

// PutObjectWithContext puts a file to S3.
func (c *client) PutObjectWithContext(ctx aws.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, error) {
	return c.s3api.PutObjectWithContext(ctx, input, opts...)
}

// GetObjectWithContext gets a file from S3.
func (c *client) GetObjectWithContext(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error) {
	return c.s3api.GetObjectWithContext(ctx, input, opts...)
}

// Storage is a storage.Storage implementation for AWS S3.
type Storage struct {
	// config is a storage config for s3.
	config *Config
	// client is an instance of S3Client interface to call API.
	// It is intended to be replaced with a mock for testing.
	client Client
}

var _ storage.Storage = (*Storage)(nil)

// NewStorage returns a new instance of Storage.
func NewStorage(config *Config, client Client) (*Storage, error) {
	if client == nil {
		var err error
		client, err = newClient(config)
		if err != nil {
			return nil, err
		}
	}

	s := &Storage{
		config: config,
		client: client,
	}

	return s, nil
}

// newClient returns a new instance of Client.
func newClient(config *Config) (Client, error) {
	cfg := &awsbase.Config{
		AccessKey:            config.AccessKey,
		AssumeRoleARN:        config.RoleARN,
		Profile:              config.Profile,
		Region:               config.Region,
		SecretKey:            config.SecretKey,
		SkipCredsValidation:  config.SkipCredentialsValidation,
		SkipMetadataApiCheck: config.SkipMetadataAPICheck,
	}

	sess, err := awsbase.GetSession(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to new s3 client: %s", err)
	}

	client := s3.New(sess.Copy(&aws.Config{
		Endpoint:         aws.String(config.Endpoint),
		S3ForcePathStyle: aws.Bool(config.ForcePathStyle),
	}))

	return client, nil
}

// Write writes migration history data to storage.
func (s *Storage) Write(ctx context.Context, b []byte) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(s.config.Key),
		Body:   bytes.NewReader(b),
	}
	if s.config.KmsKeyID != "" {
		input.SSEKMSKeyId = &s.config.KmsKeyID
		input.ServerSideEncryption = aws.String("aws:kms")
	}

	_, err := s.client.PutObjectWithContext(ctx, input)

	return err
}

// Read reads migration history data from storage.
// If the key does not exist, it is assumed to be uninitialized and returns
// an empty array instead of an error.
func (s *Storage) Read(ctx context.Context) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(s.config.Key),
	}

	output, err := s.client.GetObjectWithContext(ctx, input)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoSuchKey" {
			// If the key does not exist
			return []byte{}, nil
		}
		// unexpected error
		return nil, err
	}

	defer output.Body.Close()

	buf := bytes.NewBuffer(nil)
	_, err = buf.ReadFrom(output.Body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
