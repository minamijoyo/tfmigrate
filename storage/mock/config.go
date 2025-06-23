package mock

import "github.com/minamijoyo/tfmigrate/storage"

// Config is a config for mock storage.
type Config struct {
	// Data stores a serialized data for history.
	Data string `hcl:"data"`
	// WriteError is a flag to return an error on Write().
	WriteError bool `hcl:"write_error"`
	// ReadError is a flag to return an error on Read().
	ReadError bool `hcl:"read_error"`
	// LockExists is a flag to track if a lock file exists.
	LockExists bool `hcl:"lock_exists"`

	// A reference to an instance of mock storage for testing.
	s *Storage
}

// Config implements a storage.Config.
var _ storage.Config = (*Config)(nil)

// NewStorage returns a new instance of storage.Storage.
func (c *Config) NewStorage() (storage.Storage, error) {
	s, err := NewStorage(c)

	// store a reference for test assertion.
	c.s = s
	return s, err
}

// Storage returns a reference to mock storage for testing.
func (c *Config) Storage() *Storage {
	return c.s
}
