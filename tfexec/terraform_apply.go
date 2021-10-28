package tfexec

import (
	"context"
	"os"
)

// Apply applies changes.
// If a plan is given, use it for the input plan.
func (c *terraformCLI) Apply(ctx context.Context, plan *Plan, opts ...string) error {
	args := []string{"apply"}
	args = append(args, opts...)

	if plan != nil {
		tmpPlan, err := writeTempFile(plan.Bytes())
		defer os.Remove(tmpPlan.Name())
		if err != nil {
			return err
		}
		args = append(args, tmpPlan.Name())
	}

	_, _, err := c.Run(ctx, args...)

	return err
}
