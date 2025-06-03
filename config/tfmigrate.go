package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/zclconf/go-cty/cty"

	"github.com/minamijoyo/tfmigrate/history"
)

// ConfigurationFile represents a file for CLI settings in HCL.
type ConfigurationFile struct {
	// Tfmigrate is a top-level block.
	// It must contain only one block, and multiple blocks are not allowed.
	Tfmigrate TfmigrateBlock `hcl:"tfmigrate,block"`
}

// TfmigrateBlock represents a block for CLI settings in HCL.
type TfmigrateBlock struct {
	// MigrationDir is a path to directory where migration files are stored.
	// Default to `.` (current directory).
	MigrationDir string `hcl:"migration_dir,optional"`
	// IsBackendTerraformCloud is a boolean indicating whether a backend is
	// stored remotely in Terraform Cloud. Defaults to false.
	IsBackendTerraformCloud bool `hcl:"is_backend_terraform_cloud,optional"`
	// ExecPath is a string indicating how terraform command is executed.
	// Default to `terraform`.
	ExecPath string `hcl:"exec_path,optional"`
	// FromTfExecPath is a string indicating how terraform command is executed for the source directory.
	// Overrides ExecPath for the source directory when specified.
	FromTfExecPath string `hcl:"from_tf_exec_path,optional"`
	// ToTfExecPath is a string indicating how terraform command is executed for the destination directory.
	// Overrides ExecPath for the destination directory when specified.
	ToTfExecPath string `hcl:"to_tf_exec_path,optional"`
	// History is a block for migration history management.
	History *HistoryBlock `hcl:"history,block"`
}

// TfmigrateConfig is a config for top-level CLI settings.
// TfmigrateBlock is just used for parsing HCL and
// TfmigrateConfig is used for building application logic.
type TfmigrateConfig struct {
	// MigrationDir is a path to directory where migration files are stored.
	// Default to `.` (current directory).
	MigrationDir string
	// IsBackendTerraformCloud is a boolean representing whether the remote
	// backend is TerraformCloud. Defaults to a value of false.
	IsBackendTerraformCloud bool
	// ExecPath is a string indicating how terraform command is executed.
	// Default to `terraform`.
	ExecPath string
	// FromTfExecPath is a string indicating how terraform command is executed for the source directory.
	// Overrides ExecPath for the source directory when specified.
	FromTfExecPath string
	// ToTfExecPath is a string indicating how terraform command is executed for the destination directory.
	// Overrides ExecPath for the destination directory when specified.
	ToTfExecPath string
	// History is a config for migration history management.
	History *history.Config
}

// LoadConfigurationFile is a helper function which reads and parses a given configuration file.
func LoadConfigurationFile(filename string) (*TfmigrateConfig, error) {
	source, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return ParseConfigurationFile(filename, source)
}

// envVarMap returns a map of environment variables.
func envVarMap() cty.Value {
	envMap := make(map[string]cty.Value)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		envMap[pair[0]] = cty.StringVal(pair[1])
	}
	return cty.MapVal(envMap)
}

// ParseConfigurationFile parses a given source of configuration file and
// returns a TfmigrateConfig.
// Note that this method does not read a file and you should pass source of config in bytes.
// The filename is used for error message and selecting HCL syntax (.hcl and .json).
func ParseConfigurationFile(filename string, source []byte) (*TfmigrateConfig, error) {
	// Decode tfmigrate block.
	var f ConfigurationFile
	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"env": envVarMap(),
		},
	}
	err := hclsimple.Decode(filename, source, ctx, &f)
	if err != nil {
		return nil, fmt.Errorf("failed to decode setting file: %s, err: %s", filename, err)
	}

	config := NewDefaultConfig()
	if len(f.Tfmigrate.MigrationDir) > 0 {
		config.MigrationDir = f.Tfmigrate.MigrationDir
	}
	if f.Tfmigrate.IsBackendTerraformCloud {
		config.IsBackendTerraformCloud = f.Tfmigrate.IsBackendTerraformCloud
	}
	if len(f.Tfmigrate.ExecPath) > 0 {
		config.ExecPath = f.Tfmigrate.ExecPath
	}
	if len(f.Tfmigrate.FromTfExecPath) > 0 {
		config.FromTfExecPath = f.Tfmigrate.FromTfExecPath
	}
	if len(f.Tfmigrate.ToTfExecPath) > 0 {
		config.ToTfExecPath = f.Tfmigrate.ToTfExecPath
	}

	if f.Tfmigrate.History != nil {
		history, err := parseHistoryBlock(*f.Tfmigrate.History, ctx)
		if err != nil {
			return nil, err
		}
		config.History = history
	}

	return config, nil
}

// NewDefaultConfig returns a new instance of TfmigrateConfig.
func NewDefaultConfig() *TfmigrateConfig {
	// Get default exec path from environment variable
	execPath := os.Getenv("TFMIGRATE_EXEC_PATH")
	if len(execPath) == 0 {
		execPath = "terraform"
	}

	return &TfmigrateConfig{
		MigrationDir:            ".",
		IsBackendTerraformCloud: false,
		ExecPath:                execPath,
		FromTfExecPath:          "",
		ToTfExecPath:            "",
	}
}
