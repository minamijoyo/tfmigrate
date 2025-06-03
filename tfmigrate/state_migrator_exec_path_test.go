package tfmigrate

import (
	"testing"
)

// TestStateMigratorExecPathConfig verifies that the correct exec path is used
// based on the migrator options
func TestStateMigratorExecPathConfig(t *testing.T) {
	dir := "test-dir"

	testCases := []struct {
		name         string
		option       *MigratorOption
		expectedPath string
	}{
		{
			name:         "with default path",
			option:       &MigratorOption{},
			expectedPath: "terraform", // Default
		},
		{
			name: "with exec path",
			option: &MigratorOption{
				ExecPath: "tofu",
			},
			expectedPath: "tofu",
		},
		{
			name: "with source path",
			option: &MigratorOption{
				SourceExecPath: "terraform1",
			},
			expectedPath: "terraform1",
		},
		{
			name: "with both exec path and source path",
			option: &MigratorOption{
				ExecPath:       "tofu",
				SourceExecPath: "terraform1",
			},
			expectedPath: "terraform1", // Source path takes precedence
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a migrator with the given options
			actions := []StateAction{}
			m := NewStateMigrator(dir, "", actions, tc.option, false, false)

			// Check the terraform CLI path
			tfPath := m.tf.ExecPath()
			if tfPath != tc.expectedPath {
				t.Errorf("tf.ExecPath() = %s, want %s", tfPath, tc.expectedPath)
			}
		})
	}
}
