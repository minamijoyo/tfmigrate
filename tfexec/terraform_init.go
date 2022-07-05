package tfexec

import (
	"context"
)

// Init initializes the current work directory.
func (c *terraformCLI) Init(ctx context.Context, opts ...string) error {
	args := []string{"init"}
	args = append(args, opts...)
	_, _, err := c.Run(ctx, args...)
	return err
}
