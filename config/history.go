package config

// HistoryBlock represents a block for migration history management in HCL.
type HistoryBlock struct {
	// MigrationDir is a path to directory where migratoin files are stored.
	MigrationDir string `hcl:"migration_dir"`
	// Storage is a block for migration history data store.
	Storage StorageBlock `hcl:"storage,block"`
}

// HistoryConfig is a config for migration history management.
type HistoryConfig struct {
	// MigrationDir is a path to directory where migratoin files are stored.
	MigrationDir string
	// Storage is an interface of factory method for Storage
	Storage StorageConfig
}

// parseHistoryBlock parses a history block and returns a HistoryConfig.
func parseHistoryBlock(b HistoryBlock) (*HistoryConfig, error) {
	storage, err := parseStorageBlock(b.Storage)
	if err != nil {
		return nil, err
	}

	history := &HistoryConfig{
		MigrationDir: b.MigrationDir,
		Storage:      storage,
	}

	return history, nil
}
