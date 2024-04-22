package azure

import "github.com/minamijoyo/tfmigrate/storage"

type Config struct {
	AccessKey     string `hcl:"access_key,optional"`
	AccountName   string `hcl:"storage_account_name"`
	ContainerName string `hcl:"container_name"`
	BlobName      string `hcl:"blob_name,optional"`
}

// Config implements a storage.Config.
var _ storage.Config = (*Config)(nil)

// NewStorage returns a new instance of storage.Storage.
func (c *Config) NewStorage() (storage.Storage, error) {
	return NewStorage(c, nil)
}
