package command

import (
	"context"
	"fmt"
	"log"
	"strings"

	flag "github.com/spf13/pflag"
)

// ApplyCommand is a command which computes a new state and pushes it to the remote state.
type ApplyCommand struct {
	Meta
	backendConfig []string
}

// Run runs the procedure of this command.
func (c *ApplyCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("apply", flag.ContinueOnError)
	cmdFlags.StringVar(&c.configFile, "config", defaultConfigFile, "A path to tfmigrate config file")
	cmdFlags.StringArrayVar(&c.backendConfig, "backend-config", nil, "A backend configuration for remote state")

	if err := cmdFlags.Parse(args); err != nil {
		c.UI.Error(fmt.Sprintf("failed to parse arguments: %s", err))
		return 1
	}

	var err error
	if c.config, err = newConfig(c.configFile); err != nil {
		c.UI.Error(fmt.Sprintf("failed to load config file: %s", err))
		return 1
	}
	log.Printf("[DEBUG] [command] config: %#v\n", c.config)

	c.Option = newOption(c.config)
	c.Option.BackendConfig = c.backendConfig
	// The option may contain sensitive values such as environment variables.
	// So logging the option set log level to DEBUG instead of INFO.
	log.Printf("[DEBUG] [command] option: %#v\n", c.Option)

	if len(cmdFlags.Args()) != 1 {
		c.UI.Error(fmt.Sprintf("The command expects 1 argument, but got %d", len(cmdFlags.Args())))
		c.UI.Error(c.Help())
		return 1
	}

	// non-history mode
	if c.config.History == nil {
		migrationFile := cmdFlags.Arg(0)
		if err = c.applyWithoutHistory(migrationFile); err != nil {
			c.UI.Error(err.Error())
			return 1
		}
		return 0
	}

	// history mode

	migrationFile := ""
	if len(cmdFlags.Args()) == 1 {
		// Apply a given single migration file and save it to history.
		migrationFile = cmdFlags.Arg(0)
	}

	// Apply all unapplied pending migrations and save them to history.
	if err = c.applyWithHistory(migrationFile); err != nil {
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

// applyWithHistory is a helper function which applies all unapplied pending migrations and saves them to history.
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
  PATH                     A path of migration file
                           Required in non-history mode. Optional in history-mode.

Options:
  --config                 A path to tfmigrate config file
  --backend-config=path    A backend configuration, a path to backend configuration file or
                           key=value format backend configuraion.
                           This option is passed to terraform init when switching backend to remote.
`
	return strings.TrimSpace(helpText)
}

// Synopsis returns one-line help text.
func (c *ApplyCommand) Synopsis() string {
	return "Compute a new state and push it to remote state"
}
