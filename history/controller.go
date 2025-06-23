package history

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/minamijoyo/tfmigrate/storage"
)

// Controller manages a migration history.
type Controller struct {
	// migrationDir is a path to directory where migration files are stored.
	migrationDir string
	// migrations is a list of migration file names.
	// We simply use the file name for identification to avoid parsing all files.
	// If a migration file format changes, it doesn't make sense that parsing
	// errors occur in old format files which have been already applied.
	// The list is sorted alphabetically.
	migrations []string
	// history is a list of applied migration logs which is persisted to a storage.
	history History
	// config customizes behavior of history management.
	config Config
}

// NewController returns a new Controller instance.
func NewController(ctx context.Context, migrationDir string, config *Config) (*Controller, error) {
	log.Printf("[DEBUG] [history] load migration dir: %s\n", migrationDir)
	migrations, err := loadMigrationFileNames(migrationDir)
	if err != nil {
		return nil, err
	}

	log.Print("[DEBUG] [history] load history\n")
	h, err := loadHistory(ctx, config.Storage)
	if err != nil {
		return nil, err
	}

	c := &Controller{
		migrationDir: migrationDir,
		migrations:   migrations,
		history:      *h,
		config:       *config,
	}

	return c, nil
}

// loadMigrationDir loads a migration directory and lists migration files from local.
// The returned slice is sorted alphabetically.
func loadMigrationFileNames(dir string) ([]string, error) {
	migrations := []string{}

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		// skip a file without .hcl or .json extension.
		ext := filepath.Ext(f.Name())
		if !(ext == ".hcl" || ext == ".json") {
			continue
		}
		// skip a hidden file such as .tfmigrate.hcl or .terraform.lock.hcl.
		if strings.HasPrefix(f.Name(), ".") {
			continue
		}

		migrations = append(migrations, f.Name())
	}

	// Migrations should be applied in the order of the file name.
	sort.Strings(migrations)
	return migrations, nil
}

// loadHistory loads a history file from a storage.
// If a given history is not found, create a new one.
func loadHistory(ctx context.Context, c storage.Config) (*History, error) {
	s, err := c.NewStorage()
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] [history] read storage %#v\n", s)
	b, err := s.Read(ctx)
	if err != nil {
		return nil, err
	}
	log.Printf("[TRACE] [history] read history file: %#v\n", b)

	// If a given history is not found, s.Read returns empty bytes with no error.
	// In this case, we assume that it's the first use and create a new history.
	if len(b) == 0 {
		log.Print("[DEBUG] [history] new empty history\n")
		return newEmptyHistory(), nil
	}

	h, err := ParseHistoryFile(b)
	if err != nil {
		return nil, err
	}

	return h, nil
}

// Save persists a current state of historyFile to storage.
func (c *Controller) Save(ctx context.Context) error {
	s, err := c.config.Storage.NewStorage()
	if err != nil {
		return err
	}

	f := newFileV1(c.history)
	b, err := f.Serialize()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] [history] write storage: %#v\n", s)
	log.Printf("[TRACE] [history] write history file: %#v\n", b)
	return s.Write(ctx, b)
}

// Migrations returns a list of all migration file names.
func (c *Controller) Migrations() []string {
	return c.migrations
}

// UnappliedMigrations returns a list of migration file names which have not
// been applied yet.
func (c *Controller) UnappliedMigrations() []string {
	unapplied := []string{}
	for _, m := range c.migrations {
		if !c.history.Contains(m) {
			unapplied = append(unapplied, m)
		}
	}

	return unapplied
}

// HistoryLength returns a number of records in history.
func (c *Controller) HistoryLength() int {
	return c.history.Length()
}

// AlreadyApplied returns true if a given migration file has already been applied.
func (c *Controller) AlreadyApplied(filename string) bool {
	return c.history.Contains(filename)
}

// AddRecord adds a record to history.
// This method doesn't persist history. Call Save() to save the history.
// If appliedAt is nil, a timestamp is automatically set to time.Now().
func (c *Controller) AddRecord(filename string, migrationType string, name string, appliedAt *time.Time) {
	timestamp := appliedAt
	if timestamp == nil {
		now := time.Now()
		timestamp = &now
	}
	r := Record{
		Type:      migrationType,
		Name:      name,
		AppliedAt: *timestamp,
	}

	c.history.Add(filename, r)
}

func (c *Controller) LockState() error {
	s, err := c.config.Storage.NewStorage()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] [history] check lock state: %#v\n", s)
	return s.WriteLock(context.Background())
}

func (c *Controller) UnlockState() error {
	s, err := c.config.Storage.NewStorage()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] [history] unlock state: %#v\n", s)
	return s.Unlock(context.Background())
}
