package storage

import "context"

// Storage is an abstraction layer for migration history data store.
// As you know, this is the equivalent of Terraform's backend, but we have
// implemented it by ourselves not to depend on Terraform internals directly.
// To support multiple cloud storages, write and read operations are limited to
// simple byte operations and a domain specific logic should not be included.
type Storage interface {
	// Write writes migration history data to storage.
	Write(ctx context.Context, b []byte) error
	// Read reads migration history data from storage.
	// If the key does not exist, it is assumed to be uninitialized and returns
	// an empty array instead of an error.
	Read(ctx context.Context) ([]byte, error)

	WriteLock(ctx context.Context) error

	Unlock(ctx context.Context) error
}
