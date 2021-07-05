package tfexec

import (
	"context"
)

//WorkspaceSelect selects the workspace "workspace". The workspace needs to exist
// in order for the switch to be successful
func (c *terraformCLI) WorkspaceSelect(ctx context.Context, workspace string, dir string) error {
	args := []string{"workspace", "select"}
	if len(workspace) > 0 {
		args = append(args, workspace)
	}
	if len(dir) > 0 {
		args = append(args, dir)
	}
	_, _, err := c.Run(ctx, args...)
	return err
}
