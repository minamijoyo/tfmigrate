package config

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	storage "github.com/minamijoyo/tfmigrate-storage"
	"github.com/minamijoyo/tfmigrate-storage/gcs"
	"github.com/minamijoyo/tfmigrate-storage/local"
	"github.com/minamijoyo/tfmigrate-storage/mock"
	"github.com/minamijoyo/tfmigrate-storage/s3"
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

// parseStorageBlock parses a storage block and returns a storage.Config.
func parseStorageBlock(b StorageBlock) (storage.Config, error) {
	switch b.Type {
	case "mock": // only for testing
		return parseMockStorageBlock(b)

	case "local":
		return parseLocalStorageBlock(b)

	case "s3":
		return parseS3StorageBlock(b)

	case "gcs":
		return parseGCSStorageBlock(b)

	default:
		return nil, fmt.Errorf("unknown history storage type: %s", b.Type)
	}
}

// parseMockStorageBlock parses a storage block for mock and returns a storage.Config.
func parseMockStorageBlock(b StorageBlock) (storage.Config, error) {
	var config mock.Config
	diags := gohcl.DecodeBody(b.Remain, nil, &config)
	if diags.HasErrors() {
		return nil, diags
	}

	return &config, nil
}

// parseLocalStorageBlock parses a storage block for local and returns a storage.Config.
func parseLocalStorageBlock(b StorageBlock) (storage.Config, error) {
	var config local.Config
	diags := gohcl.DecodeBody(b.Remain, nil, &config)
	if diags.HasErrors() {
		return nil, diags
	}

	return &config, nil
}

// parseS3StorageBlock parses a storage block for s3 and returns a storage.Config.
func parseS3StorageBlock(b StorageBlock) (storage.Config, error) {
	var config s3.Config
	diags := gohcl.DecodeBody(b.Remain, nil, &config)
	if diags.HasErrors() {
		return nil, diags
	}

	return &config, nil
}

func parseGCSStorageBlock(b StorageBlock) (storage.Config, error) {
	var config gcs.Config
	diags := gohcl.DecodeBody(b.Remain, nil, &config)
	if diags.HasErrors() {
		return nil, diags
	}

	return &config, nil
}
