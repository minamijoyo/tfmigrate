package tfexec

import (
	"context"
)

//WorkspaceNew creates a new workspace
func (c *terraformCLI) WorkspaceNew(ctx context.Context, workspace string, dir string, opts ...string) error {
	args := []string{"workspace", "new"}
	args = append(args, opts...)
	if len(workspace) > 0 {
		args = append(args, workspace)
	}
	if len(dir) > 0 {
		args = append(args, dir)
	}
	_, _, err := c.Run(ctx, args...)
	return err
}
