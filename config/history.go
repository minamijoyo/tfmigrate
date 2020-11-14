package config

import "github.com/minamijoyo/tfmigrate/history"

// HistoryBlock represents a block for migration history management in HCL.
type HistoryBlock struct {
	// Storage is a block for migration history data store.
	Storage StorageBlock `hcl:"storage,block"`
}

// parseHistoryBlock parses a history block and returns a *history.Config.
func parseHistoryBlock(b HistoryBlock) (*history.Config, error) {
	storage, err := parseStorageBlock(b.Storage)
	if err != nil {
		return nil, err
	}

	history := &history.Config{
		Storage: storage,
	}

	return history, nil
}
