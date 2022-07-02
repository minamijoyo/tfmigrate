package command

import (
	"context"
	"fmt"
	"log"
	"strings"

	flag "github.com/spf13/pflag"
)

// PlanCommand is a command which computes a new state by applying state
// migration operations to a temporary state.
type PlanCommand struct {
	Meta
	backendConfig []string
	out           string
}

// Run runs the procedure of this command.
func (c *PlanCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("plan", flag.ContinueOnError)
	cmdFlags.StringVar(&c.configFile, "config", defaultConfigFile, "A path to tfmigrate config file")
	cmdFlags.StringArrayVar(&c.backendConfig, "backend-config", nil, "A backend configuration for remote state")
	cmdFlags.StringVar(&c.out, "out", "", "[Deprecated] Save a plan file after dry-run migration to the given path")

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
	c.Option.PlanOut = c.out
	// The option may contains sensitive values such as environment variables.
	// So logging the option set log level to DEBUG instead of INFO.
	log.Printf("[DEBUG] [command] option: %#v\n", c.Option)

	// The tfmigrate plan --out=tfplan option is deprecated and doesn't work with Terraform 1.1+
	// https://github.com/minamijoyo/tfmigrate/issues/62
	if c.Option.PlanOut != "" {
		log.Println("[WARN] The --out option is deprecated without replacement and it will be removed in a future release")
	}

	if c.config.History == nil {
		// non-history mode
		if len(cmdFlags.Args()) != 1 {
			c.UI.Error(fmt.Sprintf("The command expects 1 argument, but got %d", len(cmdFlags.Args())))
			c.UI.Error(c.Help())
			return 1
		}

		migrationFile := cmdFlags.Arg(0)
		if err = c.planWithoutHistory(migrationFile); err != nil {
			c.UI.Error(err.Error())
			return 1
		}

		return 0
	}

	// history mode
	if len(cmdFlags.Args()) > 1 {
		c.UI.Error(fmt.Sprintf("The command expects 0 or 1 argument, but got %d", len(cmdFlags.Args())))
		c.UI.Error(c.Help())
		return 1
	}

	migrationFile := ""
	if len(cmdFlags.Args()) == 1 {
		// plan a given single migration file.
		migrationFile = cmdFlags.Arg(0)
	}

	// Plan all unapplied pending migrations.
	if err = c.planWithHistory(migrationFile); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	return 0
}

// planWithoutHistory is a helper function which plans a given migration file without history.
func (c *PlanCommand) planWithoutHistory(filename string) error {
	fr, err := NewFileRunner(filename, c.config, c.Option)
	if err != nil {
		return err
	}

	return fr.Plan(context.Background())
}

// planWithHistory is a helper function which plans all unapplied pending migrations.
func (c *PlanCommand) planWithHistory(filename string) error {
	ctx := context.Background()
	hr, err := NewHistoryRunner(ctx, filename, c.config, c.Option)
	if err != nil {
		return err
	}

	return hr.Plan(ctx)
}

// Help returns long-form help text.
func (c *PlanCommand) Help() string {
	helpText := `
Usage: tfmigrate plan [PATH]

Plan computes a new state by applying state migration operations to a temporary state.
It will fail if terraform plan detects any diffs with the new state.

Arguments:
  PATH                     A path of migration file
                           Required in non-history mode. Optional in history-mode.

Options:
  --config                 A path to tfmigrate config file
  --backend-config=path    Configuration to be merged with what is in the
                           configuration file's 'backend' block. This can be
                           either a path to an HCL file with key/value
                           assignments (same format as terraform.tfvars) or a
                           'key=value' format, and can be specified multiple
                           times. The backend type must be in the configuration
                           itself.

  [Deprecated]
  --out=path
                           Save a plan file after dry-run migration to the given path.
                           Note that applying the plan file only affects a local state,
                           make sure to force push it to remote after terraform apply.
                           This option doesn't work with Terraform 1.1+
`
	return strings.TrimSpace(helpText)
}

// Synopsis returns one-line help text.
func (c *PlanCommand) Synopsis() string {
	return "Compute a new state"
}
