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

// newEmptyHistory initializes a new History.
func newEmptyHistory() *History {
	records := make(map[string]Record)
	return &History{
		records: records,
	}
}

// Add adds a new record to history.
// If a given filename already exists, it updates the existing record.
func (h *History) Add(filename string, r Record) {
	h.records[filename] = r
}

// Contains returns true if a given migration has been applied.
func (h *History) Contains(filename string) bool {
	_, ok := h.records[filename]
	return ok
}

// Delete deletes a record from history.
// If a given filename doesn't exist, no-op.
func (h *History) Delete(filename string) {
	delete(h.records, filename)
}

// Clear deletes all records from history.
func (h *History) Clear() {
	h.records = make(map[string]Record)
}
