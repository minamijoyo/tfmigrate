package command

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/minamijoyo/tfmigrate/config"
	"github.com/minamijoyo/tfmigrate/tfmigrate"
)

// FileRunner is a runner for a single migration file.
type FileRunner struct {
	// A path to migration file.
	filename string
	// A global configuration.
	config *config.TfmigrateConfig
	// A definition of migration.
	mc *tfmigrate.MigrationConfig
	// A migrator instance to be run.
	m tfmigrate.Migrator
}

// NewFileRunner returns a new FileRunner instance.
func NewFileRunner(filename string, config *config.TfmigrateConfig, option *tfmigrate.MigratorOption) (*FileRunner, error) {
	path := resolveMigrationFile(config.MigrationDir, filename)
	log.Printf("[INFO] [runner] load migration file: %s\n", path)
	mc, err := loadMigrationFile(path)
	if err != nil {
		return nil, err
	}

	if option != nil {
		option.IsBackendTerraformCloud = config.IsBackendTerraformCloud
	} else {
		option = &tfmigrate.MigratorOption{
			IsBackendTerraformCloud: false,
		}
	}

	m, err := mc.Migrator.NewMigrator(option)

	if err != nil {
		return nil, err
	}

	r := &FileRunner{
		filename: filename,
		config:   config,
		mc:       mc,
		m:        m,
	}

	return r, nil
}

// loadMigrationFile is a helper function which reads and parses a migration file.
func loadMigrationFile(filename string) (*tfmigrate.MigrationConfig, error) {
	source, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config, err := config.ParseMigrationFile(filename, source)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// Plan plans a single migration.
func (r *FileRunner) Plan(ctx context.Context) error {
	return r.m.Plan(ctx)
}

// Apply applies a single migration.
func (r *FileRunner) Apply(ctx context.Context) error {
	return r.m.Apply(ctx)
}

// MigrationConfig returns an instance of migration.
// This is required for metadata stored in history
func (r *FileRunner) MigrationConfig() *tfmigrate.MigrationConfig {
	return r.mc
}

// resolveMigrationFile returns a path of migration file in migration dir.
// If a given filename is absolute path, just return it as it is.
func resolveMigrationFile(migrationDir string, filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(migrationDir, filename)
}
