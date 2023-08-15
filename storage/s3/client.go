package s3

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	awsbase "github.com/hashicorp/aws-sdk-go-base"
)

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

// PutObjectWithContext puts a file to S3.
func (c *client) PutObjectWithContext(ctx aws.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, error) {
	return c.s3api.PutObjectWithContext(ctx, input, opts...)
}

// GetObjectWithContext gets a file from S3.
func (c *client) GetObjectWithContext(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error) {
	return c.s3api.GetObjectWithContext(ctx, input, opts...)
}
