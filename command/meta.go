package command

import (
	"os"

	"github.com/minamijoyo/tfmigrate/tfmigrate"
	"github.com/mitchellh/cli"
)

// Meta are the meta-options that are available on all or most commands.
type Meta struct {
	// UI is a user interface representing input and output.
	UI cli.Ui

	// Option customizes a behaviror of Migrator.
	// It is used for shared settings across Migrator instances.
	Option *tfmigrate.MigratorOption
}

func newOption() *tfmigrate.MigratorOption {
	return &tfmigrate.MigratorOption{
		ExecPath: os.Getenv("TFMIGRATE_EXEC_PATH"),
	}
}
