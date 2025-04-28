package config

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/minamijoyo/tfmigrate/history"
)

// HistoryBlock represents a block for migration history management in HCL.
type HistoryBlock struct {
	// Storage is a block for migration history data store.
	Storage StorageBlock `hcl:"storage,block"`
}

// parseHistoryBlock parses a history block and returns a *history.Config.
func parseHistoryBlock(b HistoryBlock, ctx *hcl.EvalContext) (*history.Config, error) {
	storage, err := parseStorageBlock(b.Storage, ctx)
	if err != nil {
		return nil, err
	}

	history := &history.Config{
		Storage: storage,
	}

	return history, nil
}
