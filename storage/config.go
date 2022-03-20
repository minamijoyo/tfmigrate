package storage

// Config is an interface of factory method for Storage
type Config interface {
	// NewStorage returns a new instance of Storage.
	NewStorage() (Storage, error)
}
