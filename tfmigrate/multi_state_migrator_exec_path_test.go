package tfmigrate

import (
	"testing"
)

// TestMultiStateMigratorExecPathConfig verifies that the correct exec paths are used
// based on the migrator options
func TestMultiStateMigratorExecPathConfig(t *testing.T) {
	fromDir := "from"
	toDir := "to"

	testCases := []struct {
		name             string
		option           *MigratorOption
		expectedFromPath string
		expectedToPath   string
	}{
		{
			name:             "with default paths",
			option:           &MigratorOption{},
			expectedFromPath: "terraform", // Default
			expectedToPath:   "terraform", // Default
		},
		{
			name: "with common exec path",
			option: &MigratorOption{
				ExecPath: "tofu",
			},
			expectedFromPath: "tofu",
			expectedToPath:   "tofu",
		},
		{
			name: "with source and destination paths",
			option: &MigratorOption{
				SourceExecPath:      "terraform1",
				DestinationExecPath: "terraform2",
			},
			expectedFromPath: "terraform1",
			expectedToPath:   "terraform2",
		},
		{
			name: "with common, source, and destination paths",
			option: &MigratorOption{
				ExecPath:            "terraform",
				SourceExecPath:      "terraform1",
				DestinationExecPath: "terraform2",
			},
			expectedFromPath: "terraform1",
			expectedToPath:   "terraform2",
		},
		{
			name: "with common and source path",
			option: &MigratorOption{
				ExecPath:       "terraform",
				SourceExecPath: "terraform1",
			},
			expectedFromPath: "terraform1",
			expectedToPath:   "terraform",
		},
		{
			name: "with common and destination path",
			option: &MigratorOption{
				ExecPath:            "terraform",
				DestinationExecPath: "terraform2",
			},
			expectedFromPath: "terraform",
			expectedToPath:   "terraform2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a migrator with the given options
			actions := []MultiStateAction{}
			m := NewMultiStateMigrator(fromDir, toDir, "", "", actions, tc.option, false, false, false, "")

			// Check the fromTf path
			fromPath := m.fromTf.ExecPath()
			if fromPath != tc.expectedFromPath {
				t.Errorf("fromTf.ExecPath() = %s, want %s", fromPath, tc.expectedFromPath)
			}

			// Check the toTf path
			toPath := m.toTf.ExecPath()
			if toPath != tc.expectedToPath {
				t.Errorf("toTf.ExecPath() = %s, want %s", toPath, tc.expectedToPath)
			}
		})
	}
}
