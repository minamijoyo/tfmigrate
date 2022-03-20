package local

import "github.com/minamijoyo/tfmigrate/storage"

// Config is a config for local storage.
type Config struct {
	// Path to a migration history file. Relative to the current working directory.
	Path string `hcl:"path"`
}

// Config implements a storage.Config.
var _ storage.Config = (*Config)(nil)

// NewStorage returns a new instance of storage.Storage.
func (c *Config) NewStorage() (storage.Storage, error) {
	return NewStorage(c)
}
