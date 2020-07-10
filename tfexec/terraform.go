package tfexec

import (
	"context"
	"fmt"
	"regexp"
)

// tfVersionRe is a pattern to parse outputs from terraform version.
var tfVersionRe = regexp.MustCompile(`^Terraform v(.+)\s*$`)

// State is a named type for tfstate.
// We doesn't need to parse contents of tfstate,
// but we define it as a name type to clarify TerraformCLI interface.
type State string

// TerraformCLI is an interface for executing the terraform command.
type TerraformCLI interface {
	// Verison returns a version number of Terraform.
	Version(ctx context.Context) (string, error)

	// StatePull returns the current tfstate.
	StatePull(ctx context.Context) (State, error)
}

// terraformCLI implements the TerraformCLI interface.
type terraformCLI struct {
	// Executor is an interface which executes an arbitrary command.
	Executor
}

var _ TerraformCLI = (*terraformCLI)(nil)

// NewTerraformCLI returns an implementation of the TerraformCLI interface.
func NewTerraformCLI(e Executor) TerraformCLI {
	return &terraformCLI{
		Executor: e,
	}
}

// run is a helper method for running terraform comamnd.
func (c *terraformCLI) run(ctx context.Context, args ...string) (string, error) {
	cmd, err := c.Executor.NewCommandContext(ctx, "terraform", args...)
	if err != nil {
		return "", err
	}

	err = c.Executor.Run(cmd)
	if err != nil {
		return "", err
	}

	return cmd.Stdout(), err
}

// Verison returns a version number of Terraform.
func (c *terraformCLI) Version(ctx context.Context) (string, error) {
	stdout, err := c.run(ctx, "version")
	if err != nil {
		return "", err
	}

	matched := tfVersionRe.FindStringSubmatch(stdout)
	if len(matched) != 2 {
		return "", fmt.Errorf("failed to parse terraform version: %s", stdout)
	}
	version := matched[1]
	return version, nil
}

// StatePull returns the current tfstate.
func (c *terraformCLI) StatePull(ctx context.Context) (State, error) {
	stdout, err := c.run(ctx, "state", "pull")
	if err != nil {
		return "", err
	}

	return State(stdout), nil
}
