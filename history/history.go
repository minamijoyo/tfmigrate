package history

import "time"

// History records applied migration logs.
// This is intentionally independent of the file format because we need to
// understand the old format files even if the data structure changes.
type History struct {
	// appliedLogs is a history of which migrations were applied.
	// Only success migrations are recorded.
	appliedLogs []appliedLog
}

// appliedLog represents an applied migration log.
type appliedLog struct {
	// migration is a migration file name.
	// Record only the file names not to invalidate history when the migration
	// directory is moved.
	migration string
	// appliedAt is a timestamp when the migration was applied.
	// Note that we only record it when the migration was succeed.
	appliedAt time.Time
}
