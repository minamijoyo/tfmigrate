package tfexec

import (
	"context"
	"fmt"
	"regexp"
)

// tfWorkspaceRe is a pattern to parse outputs from terraform workspace show.
var tfWorkspaceRe = regexp.MustCompile(`^[a-zA-Z\d_-]{1,90}`)

//WorkspaceShow returns the currently selected workspace
func (c *terraformCLI) WorkspaceShow(ctx context.Context) (string, error) {
	args := []string{"workspace", "show"}
	stdout, _, err := c.Run(ctx, args...)
	if err != nil {
		return "", err
	}

	matched := tfWorkspaceRe.FindString(stdout)
	if matched == "" {
		return "", fmt.Errorf("failed to parse the current workspace : %s", stdout)
	}
	return matched, nil
}
