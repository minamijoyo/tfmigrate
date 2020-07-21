package tfexec

import "context"

// Apply applies changes.
func (c *terraformCLI) Apply(ctx context.Context, dirOrPlan string, opts ...string) error {
	args := []string{"apply"}
	args = append(args, opts...)
	if len(dirOrPlan) > 0 {
		args = append(args, dirOrPlan)
	}
	_, _, err := c.Run(ctx, args...)
	return err
}
