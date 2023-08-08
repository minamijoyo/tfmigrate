package tfexec

import (
	"context"
)

// Providers prints out a tree of modules in the referenced configuration annotated with
// their provider requirements.
func (c *terraformCLI) Providers(ctx context.Context) (string, error) {
	args := []string{"providers"}

	stdout, _, err := c.Run(ctx, args...)
	if err != nil {
		return "", err
	}

	return stdout, nil
}
