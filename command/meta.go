package command

import (
	"os"

	"github.com/minamijoyo/tfmigrate/config"
	"github.com/minamijoyo/tfmigrate/tfmigrate"
	"github.com/mitchellh/cli"
)

// a default config file path
const defaultConfigFile string = ".tfmigrate.hcl"

// Meta are the meta-options that are available on all or most commands.
type Meta struct {
	// UI is a user interface representing input and output.
	UI cli.Ui

	// A path to tfmigrate config file.
	configFile string

	// a global configuration for tfmigrate.
	config *config.TfmigrateConfig

	// Option customizes a behaviror of Migrator.
	// It is used for shared settings across Migrator instances.
	Option *tfmigrate.MigratorOption
}

func newOption() *tfmigrate.MigratorOption {
	return &tfmigrate.MigratorOption{
		ExecPath: os.Getenv("TFMIGRATE_EXEC_PATH"),
	}
}
