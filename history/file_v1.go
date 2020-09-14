package history

import (
	"encoding/json"
	"time"
)

// FileV1 represents a data structure for history file format v1.
// This is redundant and almost the same as History, but defines a separate
// data structure for persistence in case of future format changes.
type FileV1 struct {
	// Version is a file format version. It is always set to 1.
	Version int `json:"version"`
	// AppliedLogs is a history of which migrations were applied.
	// Only success migrations are recorded.
	AppliedLogs []AppliedLogV1 `json:"applied_logs"`
}

// AppliedLogV1 represents an applied migration log.
type AppliedLogV1 struct {
	// Migration is a migration file name.
	// Record only the file names not to invalidate history when the migration
	// directory is moved.
	Migration string `json:"migration"`
	// AppliedAt is a timestamp when the migration was applied.
	// Note that we only record it when the migration was succeed.
	AppliedAt time.Time `json:"applied_at"`
}

// newFileV1 converts a History to a FileV1 instance.
func newFileV1(h History) *FileV1 {
	logs := []AppliedLogV1{}
	for _, l := range h.appliedLogs {
		log := newAppliedLogV1(l)
		logs = append(logs, log)
	}

	return &FileV1{
		Version:     1,
		AppliedLogs: logs,
	}
}

// newAppliedLogV1 converts an appliedLog to an AppliedLogV1 instance.
func newAppliedLogV1(l appliedLog) AppliedLogV1 {
	return AppliedLogV1{
		Migration: l.migration,
		AppliedAt: l.appliedAt,
	}
}

// Serialize encodes a FileV1 instance to bytes.
func (f *FileV1) Serialize() ([]byte, error) {
	return json.MarshalIndent(f, "", "    ")
}

// parseHistoryFileV1 parses bytes and reteurns a History instance.
func parseHistoryFileV1(b []byte) (*History, error) {
	var f FileV1

	err := json.Unmarshal(b, &f)
	if err != nil {
		return nil, err
	}

	h := f.toHistory()

	return &h, nil
}

// toHistory converts a FileV1 to a History instance.
func (f *FileV1) toHistory() History {
	logs := []appliedLog{}
	for _, l := range f.AppliedLogs {
		logs = append(logs, l.toAppliedLog())
	}

	return History{
		appliedLogs: logs,
	}
}

// toAppliedLog converts an AppliedLogV1 to an appliedLog instance.
func (l AppliedLogV1) toAppliedLog() appliedLog {
	return appliedLog{
		migration: l.Migration,
		appliedAt: l.AppliedAt,
	}
}
