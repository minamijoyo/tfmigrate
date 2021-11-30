package command

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/minamijoyo/tfmigrate/config"
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
	ctx := context.Background()
	out, err := listMigrations(ctx, c.config, c.status)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	c.UI.Output(out)
	return 0
}

// listMigrations lists migrations.
func listMigrations(ctx context.Context, config *config.TfmigrateConfig, status string) (string, error) {
	hc, err := history.NewController(ctx, config.MigrationDir, config.History)
	if err != nil {
		return "", err
	}

	var migrations []string
	switch status {
	case "all":
		migrations = hc.Migrations()

	case "unapplied":
		migrations = hc.UnappliedMigrations()

	default:
		return "", fmt.Errorf("unknown filter for status: %s", status)
	}

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
