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

	// Option customizes a behavior of Migrator.
	// It is used for shared settings across Migrator instances.
	Option *tfmigrate.MigratorOption
}

// newConfig loads the configuration file.
func newConfig(filename string) (*config.TfmigrateConfig, error) {
	pathToLoad := filename // Start with the provided filename

	// If the provided filename is the default, check the environment variable.
	if filename == defaultConfigFile {
		envConfig := os.Getenv("TFMIGRATE_CONFIG")
		if envConfig != "" {
			// If TFMIGRATE_CONFIG is set, use its value.
			pathToLoad = envConfig
		} else {
			// If TFMIGRATE_CONFIG is not set, check if the default file exists.
			if _, err := os.Stat(defaultConfigFile); os.IsNotExist(err) {
				// If defaultConfigFile doesn't exist, return a default config.
				return config.NewDefaultConfig(), nil
			}
			// If defaultConfigFile exists, pathToLoad remains defaultConfigFile.
		}
	}

	log.Printf("[DEBUG] [command] load configuration file: %s\n", pathToLoad)
	return config.LoadConfigurationFile(pathToLoad)
}

func newOption(cfg *config.TfmigrateConfig) *tfmigrate.MigratorOption {
	// By default, use the environment variable
	execPath := os.Getenv("TFMIGRATE_EXEC_PATH")

	// If config has a value for ExecPath, use it instead of env var
	if cfg != nil && cfg.ExecPath != "" {
		execPath = cfg.ExecPath
	}

	// Create option with the base exec path
	option := &tfmigrate.MigratorOption{
		ExecPath: execPath,
	}

	// If config has values for the source/destination paths, use them
	if cfg != nil {
		if cfg.FromTfExecPath != "" {
			option.SourceExecPath = cfg.FromTfExecPath
		}
		if cfg.ToTfExecPath != "" {
			option.DestinationExecPath = cfg.ToTfExecPath
		}
		// Set IsBackendTerraformCloud from config
		option.IsBackendTerraformCloud = cfg.IsBackendTerraformCloud
	}

	return option
}
