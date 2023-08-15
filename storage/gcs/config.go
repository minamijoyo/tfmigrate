package gcs

import "github.com/minamijoyo/tfmigrate/storage"

// Config is a config for Google Cloud Storage.
// This is expected to have almost the same options as Terraform gcs backend.
// https://www.terraform.io/language/settings/backends/gcs
// However, it has many minor options and it's a pain to test all options from
// first, so we added only options we need for now.
type Config struct {
	// The name of the GCS bucket.
	Bucket string `hcl:"bucket"`
	// Path to the migration history file.
	Name string `hcl:"name"`
}

// Config implements a storage.Config.
var _ storage.Config = (*Config)(nil)

// NewStorage returns a new instance of storage.Storage.
func (c *Config) NewStorage() (storage.Storage, error) {
	return NewStorage(c, nil)
}
