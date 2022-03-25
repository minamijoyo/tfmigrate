package history

import (
	storage "github.com/minamijoyo/tfmigrate-storage"
)

// Config is a set of configurations for migration history management.
type Config struct {
	// MigrationDir is a path to directory where migration files are stored.
	MigrationDir string
	// Storage is an interface of factory method for Storage
	Storage storage.Config
}
