package command

import (
	"log"
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
	// TODO: Merge option into config
	Option *tfmigrate.MigratorOption
}

func newConfig(filename string) (*config.TfmigrateConfig, error) {
	if filename == defaultConfigFile {
		if _, err := os.Stat(defaultConfigFile); os.IsNotExist(err) {
			// If defaultConfigFile doesn't exist,
			// Ignore the error and just return a default config.
			return config.NewDefaultConfig(), nil
		}
	}

	log.Printf("[DEBUG] [command] load configuration file: %s\n", filename)
	return config.LoadConfigurationFile(filename)
}

func newOption() *tfmigrate.MigratorOption {
	return &tfmigrate.MigratorOption{
		ExecPath: os.Getenv("TFMIGRATE_EXEC_PATH"),
	}
}
