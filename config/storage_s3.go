package config

import (
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/minamijoyo/tfmigrate/history"
)

// S3StorageConfig is a config for s3 storage.
type S3StorageConfig struct {
	// Bucket is a name of s3 bucket.
	Bucket string `hcl:"bucket"`
	// Key is a path to a migration history file.
	Key string `hcl:"key"`
}

// S3StorageConfig implements a StorageConfig.
var _ StorageConfig = (*S3StorageConfig)(nil)

// NewStorage returns a new instance of S3Storage.
func (c *S3StorageConfig) NewStorage() (history.Storage, error) {
	s, err := history.NewS3Storage(c.Bucket, c.Key, nil)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// parseS3StorageBlock parses a storage block for s3 and returns a StorageConfig.
func parseS3StorageBlock(b StorageBlock) (StorageConfig, error) {
	var storage S3StorageConfig
	diags := gohcl.DecodeBody(b.Remain, nil, &storage)
	if diags.HasErrors() {
		return nil, diags
	}

	return &storage, nil
}
