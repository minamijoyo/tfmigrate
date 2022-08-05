package command

import (
	"os"
	"path/filepath"
	"testing"
)

// setupMigrationFile is a test helper for setting up a temporary migration file.
// It returns a path of migration file.
func setupMigrationFile(t *testing.T, source string) string {
	migrationDir, err := os.MkdirTemp("", "migrationDir")
	if err != nil {
		t.Fatalf("failed to craete migration dir: %s", err)
	}
	t.Cleanup(func() { os.RemoveAll(migrationDir) })

	path := filepath.Join(migrationDir, "test.hcl")
	err = os.WriteFile(path, []byte(source), 0600)
	if err != nil {
		t.Fatalf("failed to write migration file: %s", err)
	}

	return path
}

// setupMigrationDir is a test helper for setting up a temporary migration dir.
// A given migrations is a map of filename to source of migration file.
// It returns a path of migration dir.
func setupMigrationDir(t *testing.T, migrations map[string]string) string {
	migrationDir, err := os.MkdirTemp("", "migrationDir")
	if err != nil {
		t.Fatalf("failed to craete migration dir: %s", err)
	}
	t.Cleanup(func() { os.RemoveAll(migrationDir) })

	for filename, source := range migrations {
		path := filepath.Join(migrationDir, filename)
		err = os.WriteFile(path, []byte(source), 0600)
		if err != nil {
			t.Fatalf("failed to write migration file: %s", err)
		}
	}

	return migrationDir
}
