package history

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"sort"
)

// Controller manages a migration history.
type Controller struct {
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
func NewController(ctx context.Context, config Config) (*Controller, error) {
	migrations, err := loadMigrationFileNames(config.MigrationDir)
	if err != nil {
		return nil, err
	}

	h, err := loadHistory(ctx, config.Storage)
	if err != nil {
		return nil, err
	}

	c := &Controller{
		migrations: migrations,
		history:    *h,
		config:     config,
	}

	return c, nil
}

// loadMigrationDir loads a migration directory and lists migration files from local.
// The returned slice is sorted alphabetically.
func loadMigrationFileNames(dir string) ([]string, error) {
	migrations := []string{}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		// skip a file without .hcl or .json extension.
		ext := filepath.Ext(f.Name())
		if !(ext == ".hcl" || ext == ".json") {
			continue
		}

		migrations = append(migrations, f.Name())
	}

	// Migrations should be applied in the order of the file name.
	sort.Strings(migrations)
	return migrations, nil
}

// loadHistory loads a history file from a storage.
func loadHistory(ctx context.Context, c StorageConfig) (*History, error) {
	s, err := c.NewStorage()
	if err != nil {
		return nil, err
	}

	b, err := s.Read(ctx)
	if err != nil {
		return nil, err
	}

	h, err := parseHistoryFile(b)
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

	return s.Write(ctx, b)
}
