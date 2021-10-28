package tfexec

import "context"

// Destroy destroys resources.
func (c *terraformCLI) Destroy(ctx context.Context, opts ...string) error {
	args := []string{"destroy"}
	args = append(args, opts...)
	_, _, err := c.Run(ctx, args...)
	return err
}
