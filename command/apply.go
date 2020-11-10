package command

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/minamijoyo/tfmigrate/config"
	flag "github.com/spf13/pflag"
)

// ApplyCommand is a command which computes a new state and pushes it to remote state.
type ApplyCommand struct {
	Meta
}

// Run runs the procedure of this command.
func (c *ApplyCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("apply", flag.ContinueOnError)
	cmdFlags.StringVar(&c.configFile, "config", defaultConfigFile, "A path to tfmigrate config file")
	if err := cmdFlags.Parse(args); err != nil {
		c.UI.Error(fmt.Sprintf("failed to parse arguments: %s", err))
		return 1
	}

	c.config = config.NewDefaultConfig()
	if len(c.configFile) != 0 {
		config, err := config.LoadConfigurationFile(c.configFile)
		if err != nil {
			c.UI.Error(fmt.Sprintf("failed to load config file: %s", err))
			return 1
		}
		c.config = config
	}

	c.Option = newOption()
	// The option may contains sensitive values such as environment variables.
	// So logging the option set log level to DEBUG instead of INFO.
	log.Printf("[DEBUG] [command] option: %#v\n", c.Option)

	if c.config.History == nil {
		// non-history mode
		if len(cmdFlags.Args()) != 1 {
			c.UI.Error(fmt.Sprintf("The command expects 1 argument, but got %d", len(cmdFlags.Args())))
			c.UI.Error(c.Help())
			return 1
		}

		migrationFile := cmdFlags.Arg(0)
		err := c.applyWithoutHistory(migrationFile)
		if err != nil {
			c.UI.Error(err.Error())
			return 1
		}
	}

	// history mode
	if len(cmdFlags.Args()) > 1 {
		c.UI.Error(fmt.Sprintf("The command expects 0 or 1 argument, but got %d", len(cmdFlags.Args())))
		c.UI.Error(c.Help())
		return 1
	}

	migrationFile := ""
	if len(cmdFlags.Args()) == 1 {
		// Apply a given single migration file and save it to history.
		migrationFile = cmdFlags.Arg(0)
	}

	// Apply all unapplied pending migrations and save them to history.
	err := c.applyWithHistory(migrationFile)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	return 0
}

// applyWithoutHistory is a helper function which applies a given migration file without history.
func (c *ApplyCommand) applyWithoutHistory(filename string) error {
	fr, err := NewFileRunner(filename, c.config, c.Option)
	if err != nil {
		return err
	}

	return fr.Apply(context.Background())
}

// applyWithHistory is a helper function which applies all unapplied pending migrations and save them to history.
func (c *ApplyCommand) applyWithHistory(filename string) error {
	ctx := context.Background()
	hr, err := NewHistoryRunner(ctx, filename, c.config, c.Option)
	if err != nil {
		return err
	}

	return hr.Apply(ctx)
}

// Help returns long-form help text.
func (c *ApplyCommand) Help() string {
	helpText := `
Usage: tfmigrate apply [PATH]

Apply computes a new state and pushes it to remote state.
It will fail if terraform plan detects any diffs with the new state.

Arguments
  PATH               A path of migration file
                     Required in non-history mode. Optinal in history-mode.

Options:
  --config           A path to tfmigrate config file
`
	return strings.TrimSpace(helpText)
}

// Synopsis returns one-line help text.
func (c *ApplyCommand) Synopsis() string {
	return "Compute a new state and push it to remote state"
}
