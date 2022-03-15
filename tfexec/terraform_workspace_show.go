package tfexec

import (
	"context"
	"strings"
)

// WorkspaceShow returns the currently selected workspace
func (c *terraformCLI) WorkspaceShow(ctx context.Context) (string, error) {
	args := []string{"workspace", "show"}
	stdout, _, err := c.Run(ctx, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(stdout, "\n"), nil
}
