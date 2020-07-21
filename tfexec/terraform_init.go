package tfexec

import "context"

// Init initializes a given work directory.
func (c *terraformCLI) Init(ctx context.Context, dir string, opts ...string) error {
	args := []string{"init"}
	args = append(args, opts...)
	if len(dir) > 0 {
		args = append(args, dir)
	}
	_, _, err := c.Run(ctx, args...)
	return err
}
