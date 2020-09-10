package config

import (
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/minamijoyo/tfmigrate/history"
)

// LocalStorageConfig is a config for local storage.
type LocalStorageConfig struct {
	// Path to a migration history file. Relative to the current working directory.
	Path string `hcl:"path"`
}

// LocalStorageConfig implements a StorageConfig.
var _ StorageConfig = (*LocalStorageConfig)(nil)

// NewStorage returns a new instance of LocalStorage.
func (c *LocalStorageConfig) NewStorage() (history.Storage, error) {
	s := history.NewLocalStorage(c.Path)
	return s, nil
}

// parseLocalStorageBlock parses a storage block for local and returns a StorageConfig.
func parseLocalStorageBlock(b StorageBlock) (StorageConfig, error) {
	var storage LocalStorageConfig
	diags := gohcl.DecodeBody(b.Remain, nil, &storage)
	if diags.HasErrors() {
		return nil, diags
	}

	return &storage, nil
}
