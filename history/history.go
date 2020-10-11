package history

import "time"

// History records applied migration logs.
type History struct {
	// records is a set of applied migration log.
	// Only success migrations are recorded.
	// A key is migration file name.
	// We record only the file name not to invalidate history when the migration
	// directory is moved.
	records map[string]Record
}

// Record represents an applied migration log.
type Record struct {
	// Type is a migration type.
	Type string
	// Name is a migration name.
	Name string
	// AppliedAt is a timestamp when the migration was applied.
	// Note that we only record it when the migration was succeed.
	AppliedAt time.Time
}
