package command

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/minamijoyo/tfmigrate/history"
	flag "github.com/spf13/pflag"
)

// ListCommand is a command which lists migrations.
type ListCommand struct {
	Meta
	status string
}

// Run runs the procedure of this command.
func (c *ListCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("list", flag.ContinueOnError)
	cmdFlags.StringVar(&c.configFile, "config", defaultConfigFile, "A path to tfmigrate config file")
	cmdFlags.StringVar(&c.status, "status", "all", "A filter for migration status")

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

	c.Option = newOption()
	// The option may contains sensitive values such as environment variables.
	// So logging the option set log level to DEBUG instead of INFO.
	log.Printf("[DEBUG] [command] option: %#v\n", c.Option)

	if c.config.History == nil {
		// non-history mode
		c.UI.Error("no history setting")
		return 1
	}

	// history mode
	switch c.status {
	case "all":
		out, err := c.listMigrations()
		if err != nil {
			c.UI.Error(err.Error())
			return 1
		}
		c.UI.Output(out)
		return 0

	case "unapplied":
		out, err := c.listUnapplied()
		if err != nil {
			c.UI.Error(err.Error())
			return 1
		}
		c.UI.Output(out)
		return 0

	default:
		c.UI.Error(fmt.Sprintf("unknown filter for status: %s", c.status))
		return 1
	}
}

// listMigrations lists all migrations.
func (c *ListCommand) listMigrations() (string, error) {
	ctx := context.Background()
	hc, err := history.NewController(ctx, c.config.MigrationDir, c.config.History)
	if err != nil {
		return "", err
	}

	migrations := hc.Migrations()
	out := strings.Join(migrations, "\n")

	return out, nil
}

// listUnapplied lists unapplied migrations.
func (c *ListCommand) listUnapplied() (string, error) {
	ctx := context.Background()
	hc, err := history.NewController(ctx, c.config.MigrationDir, c.config.History)
	if err != nil {
		return "", err
	}

	migrations := hc.UnappliedMigrations()
	out := strings.Join(migrations, "\n")

	return out, nil
}

// Help returns long-form help text.
func (c *ListCommand) Help() string {
	helpText := `
Usage: tfmigrate list

List migrations.

Options:
  --config           A path to tfmigrate config file
  --status           A filter for migration status
                     Valid values are as follows:
                       - all (default)
                       - unapplied
`
	return strings.TrimSpace(helpText)
}

// Synopsis returns one-line help text.
func (c *ListCommand) Synopsis() string {
	return "List migrations"
}
