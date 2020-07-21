package tfexec

import (
	"context"
	"fmt"
	"os"
)

// Apply applies changes.
// If a plan is given, use it for the input plan.
func (c *terraformCLI) Apply(ctx context.Context, plan *Plan, dir string, opts ...string) error {
	args := []string{"apply"}
	args = append(args, opts...)

	if plan != nil {
		if len(dir) > 0 {
			return fmt.Errorf("failed to build options. The plan argument (!= nil) and the dir argument cannot be set at the same time: plan=%v, dir=%v", plan, dir)
		}
		tmpPlan, err := writeTempFile(plan.Bytes())
		defer os.Remove(tmpPlan.Name())
		if err != nil {
			return err
		}
		args = append(args, tmpPlan.Name())
	}

	if len(dir) > 0 {
		args = append(args, dir)
	}

	_, _, err := c.Run(ctx, args...)

	return err
}
