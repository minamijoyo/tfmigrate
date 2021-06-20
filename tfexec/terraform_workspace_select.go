package tfexec

import (
	"context"
)

//WorkspaceSelect selects the workspace "workspace". The workspace needs to exist
// in order for the switch to be successful
func (c *terraformCLI) WorkspaceSelect(ctx context.Context, workspace string, opts ...string) error {
	args := []string{"workspace", "select"}
	args = append(args, opts...)
	if len(workspace) > 0 {
		args = append(args, workspace)
	}
	_, _, err := c.Run(ctx, args...)
	return err
}
