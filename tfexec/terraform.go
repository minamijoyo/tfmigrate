package tfexec

import (
	"context"
	"fmt"
	"regexp"
)

// tfVersionRe is a pattern to parse outputs from terraform version.
var tfVersionRe = regexp.MustCompile(`^Terraform v(.+)\s*$`)

// TerraformCLI is an interface for executing the terraform command.
type TerraformCLI interface {
	// Verison returns a version number of Terraform.
	Version(ctx context.Context) (string, error)

	// StatePull returns contents of tfstate.
	StatePull(ctx context.Context) (string, error)
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

// StatePull returns contents of tfstate.
func (c *terraformCLI) StatePull(ctx context.Context) (string, error) {
	stdout, err := c.run(ctx, "state", "pull")
	if err != nil {
		return "", err
	}

	return stdout, nil
}
