package command

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/minamijoyo/tfmigrate/config"
)

// PlanCommand is a command which computes a new state by applying state
// migration operations to a temporary state.
type PlanCommand struct {
	Meta
	path string
}

// Run runs the procedure of this command.
func (c *PlanCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("plan", flag.ContinueOnError)
	if err := cmdFlags.Parse(args); err != nil {
		c.UI.Error(fmt.Sprintf("failed to parse arguments: %s", err))
		return 1
	}

	if len(cmdFlags.Args()) != 1 {
		c.UI.Error(fmt.Sprintf("The command expects 1 argument, but got %d", len(cmdFlags.Args())))
		c.UI.Error(c.Help())
		return 1
	}

	c.Option = newOption()
	c.path = cmdFlags.Arg(0)

	source, err := ioutil.ReadFile(c.path)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	config, err := config.ParseMigrationFile(c.path, source)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	m, err := config.NewMigrator(c.Option)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	err = m.Plan(context.Background())
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	return 0
}

// Help returns long-form help text.
func (c *PlanCommand) Help() string {
	helpText := `
Usage: tfmigrate plan <PATH>

Plan computes a new state by applying state migration operations to a temporary state.
It will fail if terraform plan detects any diffs with the new state.

Arguments
  PATH               A path of migration file
`
	return strings.TrimSpace(helpText)
}

// Synopsis returns one-line help text.
func (c *PlanCommand) Synopsis() string {
	return "Computes a new state"
}
