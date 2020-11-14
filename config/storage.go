package config

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/minamijoyo/tfmigrate/history"
)

// StorageBlock represents a block for migration history data store in HCL.
type StorageBlock struct {
	// Type is a type for storage.
	// Valid values are as follows:
	// - mock
	// - local
	// - s3
	Type string `hcl:"type,label"`
	// Remain is a body of storage block.
	// We first decode only a block header and then decode schema depending on
	// its type label.
	Remain hcl.Body `hcl:",remain"`
}

// parseStorageBlock parses a storage block and returns a history.StorageConfig.
func parseStorageBlock(b StorageBlock) (history.StorageConfig, error) {
	switch b.Type {
	case "mock": // only for testing
		return parseMockStorageBlock(b)

	case "local":
		return parseLocalStorageBlock(b)

	case "s3":
		return parseS3StorageBlock(b)

	default:
		return nil, fmt.Errorf("unknown history storage type: %s", b.Type)
	}
}

// parseMockStorageBlock parses a storage block for mock and returns a history.StorageConfig.
func parseMockStorageBlock(b StorageBlock) (history.StorageConfig, error) {
	var config history.MockStorageConfig
	diags := gohcl.DecodeBody(b.Remain, nil, &config)
	if diags.HasErrors() {
		return nil, diags
	}

	return &config, nil
}

// parseLocalStorageBlock parses a storage block for local and returns a history.StorageConfig.
func parseLocalStorageBlock(b StorageBlock) (history.StorageConfig, error) {
	var config history.LocalStorageConfig
	diags := gohcl.DecodeBody(b.Remain, nil, &config)
	if diags.HasErrors() {
		return nil, diags
	}

	return &config, nil
}

// parseS3StorageBlock parses a storage block for s3 and returns a history.StorageConfig.
func parseS3StorageBlock(b StorageBlock) (history.StorageConfig, error) {
	var config history.S3StorageConfig
	diags := gohcl.DecodeBody(b.Remain, nil, &config)
	if diags.HasErrors() {
		return nil, diags
	}

	return &config, nil
}
