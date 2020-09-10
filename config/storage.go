package config

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/minamijoyo/tfmigrate/history"
)

// StorageBlock represents a block for migration history data store in HCL.
type StorageBlock struct {
	// Type is a type for storage.
	// Valid values are as follows:
	// - local
	// - s3
	Type string `hcl:"type,label"`
	// Remain is a body of storage block.
	// We first decode only a block header and then decode schema depending on
	// its type label.
	Remain hcl.Body `hcl:",remain"`
}

// StorageConfig is an interface of factory method for Storage
type StorageConfig interface {
	// NewStorage returns a new instance of Storage.
	NewStorage() (history.Storage, error)
}

// parseStorageBlock parses a storage block and returns StorageConfig.
func parseStorageBlock(b StorageBlock) (StorageConfig, error) {
	switch b.Type {
	case "local":
		storage, err := parseLocalStorageBlock(b)
		if err != nil {
			return nil, err
		}
		return storage, nil

	case "s3":
		storage, err := parseS3StorageBlock(b)
		if err != nil {
			return nil, err
		}
		return storage, nil

	default:
		return nil, fmt.Errorf("unknown history storage type: %s", b.Type)
	}
}
