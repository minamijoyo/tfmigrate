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
	// Records is a set of applied migration log.
	// Only success migrations are recorded.
	// A key is migration file name.
	// We record only the file name not to invalidate history when the migration
	// directory is moved.
	Records map[string]RecordV1 `json:"records"`
}

// RecordV1 represents an applied migration log.
type RecordV1 struct {
	// Type is a migration type.
	Type string `json:"type"`
	// Name is a migration name.
	Name string `json:"name"`
	// AppliedAt is a timestamp when the migration was applied.
	// Note that we only record it when the migration was succeed.
	AppliedAt time.Time `json:"applied_at"`
}

// newFileV1 converts a History to a FileV1 instance.
func newFileV1(h History) *FileV1 {
	m := make(map[string]RecordV1)
	for k, v := range h.records {
		r := newRecordV1(v)
		m[k] = r
	}

	return &FileV1{
		Version: 1,
		Records: m,
	}
}

// newRecordV1 converts a Record to a RecordV1 instance.
func newRecordV1(r Record) RecordV1 {
	// nolint gosimple
	return RecordV1{
		Type:      r.Type,
		Name:      r.Name,
		AppliedAt: r.AppliedAt,
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
	m := make(map[string]Record)
	for k, v := range f.Records {
		r := v.toRecord()
		m[k] = r
	}
	return History{
		records: m,
	}
}

// toRecord converts a RecordV1 to a Record instance.
func (r RecordV1) toRecord() Record {
	// nolint gosimple
	return Record{
		Type:      r.Type,
		Name:      r.Name,
		AppliedAt: r.AppliedAt,
	}
}
