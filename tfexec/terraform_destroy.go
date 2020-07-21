package tfexec

import "context"

// Destroy destroys resources.
func (c *terraformCLI) Destroy(ctx context.Context, dir string, opts ...string) error {
	args := []string{"destroy"}
	args = append(args, opts...)
	if len(dir) > 0 {
		args = append(args, dir)
	}
	_, _, err := c.Run(ctx, args...)
	return err
}
