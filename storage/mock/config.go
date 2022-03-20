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

	// A reference to an instance of mock storage for testing.
	s *Storage
}

// Config implements a storage.Config.
var _ storage.Config = (*Config)(nil)

// NewStorage returns a new instance of storage.Storage.
func (c *Config) NewStorage() (storage.Storage, error) {
	s := NewStorage(c.Data, c.WriteError, c.ReadError)

	// store a reference for test assertion.
	c.s = s
	return s, nil
}

// StorageData returns a raw data in mock storage for testing.
func (c *Config) StorageData() string {
	return c.s.data
}
