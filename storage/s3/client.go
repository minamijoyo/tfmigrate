package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awsbase "github.com/hashicorp/aws-sdk-go-base/v2"
)

// Client is an abstraction layer for AWS S3 API.
// It is intended to be replaced with a mock for testing.
type Client interface {
	// PutObject puts a file to S3.
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	// GetObject gets a file from S3.
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)

	// Note: Additional methods can be added as needed.
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

// client is a real implementation of the Client.
type client struct {
	s3Client  *s3.Client
	awsConfig aws.Config
}

// newClient returns a new instance of Client.
func newClient(config *Config) (Client, error) {
	cfg := &awsbase.Config{
		AccessKey:           config.AccessKey,
		Profile:             config.Profile,
		Region:              config.Region,
		SecretKey:           config.SecretKey,
		SkipCredsValidation: config.SkipCredentialsValidation,
	}

	if config.RoleARN != "" {
		cfg.AssumeRole = &awsbase.AssumeRole{
			RoleARN: config.RoleARN,
		}
	}

	if config.SkipMetadataAPICheck {
		cfg.EC2MetadataServiceEnableState = imds.ClientDisabled
	} else {
		cfg.EC2MetadataServiceEnableState = imds.ClientEnabled
	}

	ctx := context.Background()
	_, awsConfig, awsDiags := awsbase.GetAwsConfig(ctx, cfg)
	if awsDiags.HasError() {
		return nil, fmt.Errorf("failed to load aws config: %#v", awsDiags)
	}

	s3Client := s3.NewFromConfig(awsConfig, func(options *s3.Options) {
		if config.Endpoint != "" {
			options.BaseEndpoint = aws.String(config.Endpoint)
		}
		if config.ForcePathStyle {
			options.UsePathStyle = config.ForcePathStyle
		}
	})

	return &client{
		s3Client:  s3Client,
		awsConfig: awsConfig,
	}, nil
}

// PutObject puts a file to S3.
func (c *client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return c.s3Client.PutObject(ctx, params, optFns...)
}

// GetObject gets a file from S3.
func (c *client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return c.s3Client.GetObject(ctx, params, optFns...)
}

func (c *client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	return c.s3Client.DeleteObject(ctx, params, optFns...)
}
