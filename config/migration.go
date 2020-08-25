package config

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/minamijoyo/tfmigrate/tfmigrate"
)

// MigrationFile represents a config for migration written in HCL.
type MigrationFile struct {
	// Migration is a migration block.
	// It must contain only one block, and multiple blocks are not allowed,
	// because it's hard to re-run the file if partially failed.
	Migration MigrationBlock `hcl:"migration,block"`
}

// MigrationBlock represents a migration block in HCL.
type MigrationBlock struct {
	// Type is a type for migration.
	// Valid values are `state` and `multi_state`.
	Type string `hcl:"type,label"`
	// Name is an arbitrary name for migration.
	Name string `hcl:"name,label"`
	// Remain is a body of migration block.
	// We first decode only a block header and then decode schema depending on
	// its type label.
	Remain hcl.Body `hcl:",remain"`
}

// StateMigratorConfig is a config for StateMigrator.
type StateMigratorConfig struct {
	// Dir is a working directory for executing terraform command.
	// Default to `.` (current directory).
	Dir string `hcl:"dir,optional"`
	// Actions is a list of state action.
	// action is a plain text for state operation.
	// Valid formats are the following.
	// "mv <source> <destination>"
	// "rm <addresses>...
	// "import <address> <id>"
	// We could define strict block schema for action, but intentionally use a
	// schema-less string to allow us to easily copy terraform state command to
	// action.
	Actions []string `hcl:"actions"`
}

// MultiStateMigratorConfig is a config for MultiStateMigrator.
type MultiStateMigratorConfig struct {
	// FromDir is a working directory where states of resources move from.
	FromDir string `hcl:"from_dir"`
	// ToDir is a working directory where states of resources move to.
	ToDir string `hcl:"to_dir"`
	// Actions is a list of multi state action.
	// action is a plain text for state operation.
	// Valid formats are the following.
	// "mv <source> <destination>"
	Actions []string `hcl:"actions"`
}

// MigratorConfig is an interface of factory method for Migrator.
type MigratorConfig interface {
	// NewMigrator returns a new instance of Migrator.
	NewMigrator(o *tfmigrate.MigratorOption) (tfmigrate.Migrator, error)
}

// StateMigratorConfig implements a MigratorConfig.
var _ MigratorConfig = (*StateMigratorConfig)(nil)

// MultiStateMigratorConfig implements a MigratorConfig.
var _ MigratorConfig = (*MultiStateMigratorConfig)(nil)

// NewMigrator returns a new instance of StateMigrator.
func (c *StateMigratorConfig) NewMigrator(o *tfmigrate.MigratorOption) (tfmigrate.Migrator, error) {
	// default working directory
	dir := "."
	if len(c.Dir) > 0 {
		dir = c.Dir
	}

	if len(c.Actions) == 0 {
		return nil, fmt.Errorf("faild to NewMigrator with no actions")
	}

	// build actions from config.
	actions := []tfmigrate.StateAction{}
	for _, cmdStr := range c.Actions {
		action, err := tfmigrate.NewStateActionFromString(cmdStr)
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}

	return tfmigrate.NewStateMigrator(dir, actions, o), nil
}

// NewMigrator returns a new instance of MultiStateMigrator.
func (c *MultiStateMigratorConfig) NewMigrator(o *tfmigrate.MigratorOption) (tfmigrate.Migrator, error) {
	if len(c.Actions) == 0 {
		return nil, fmt.Errorf("faild to NewMigrator with no actions")
	}

	// build actions from config.
	actions := []tfmigrate.MultiStateAction{}
	for _, cmdStr := range c.Actions {
		action, err := tfmigrate.NewMultiStateActionFromString(cmdStr)
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}

	return tfmigrate.NewMultiStateMigrator(c.FromDir, c.ToDir, actions, o), nil
}

// ParseMigrationFile parses a given source of migration file and returns a MigratorConfig.
// Note that this method does not read a file and you should pass source of config in bytes.
// The filename is used for error message and selecting HCL syntax (.hcl and .hcl.json).
func ParseMigrationFile(filename string, source []byte) (MigratorConfig, error) {
	// Decode migration block header.
	var f MigrationFile
	err := hclsimple.Decode(filename, source, nil, &f)
	if err != nil {
		return nil, fmt.Errorf("failed to decode migration file: %s, err: %s", filename, err)
	}

	// switch schema by migration type.
	var config MigratorConfig
	m := f.Migration
	switch m.Type {
	case "state":
		var state StateMigratorConfig
		diags := gohcl.DecodeBody(m.Remain, nil, &state)
		if diags.HasErrors() {
			return nil, diags
		}
		config = &state
	case "multi_state":
		var multiState MultiStateMigratorConfig
		diags := gohcl.DecodeBody(m.Remain, nil, &multiState)
		if diags.HasErrors() {
			return nil, diags
		}
		config = &multiState
	default:
		return nil, fmt.Errorf("unknown migration type in %s: %s", filename, m.Type)
	}
	return config, nil
}
