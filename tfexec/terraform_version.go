package tfexec

import (
	"context"
	"fmt"
	"regexp"
)

// tfVersionRe is a pattern to parse outputs from terraform version.
var tfVersionRe = regexp.MustCompile(`^Terraform v(.+)\s*$`)

// Verison returns a version number of Terraform.
func (c *terraformCLI) Version(ctx context.Context) (string, error) {
	stdout, _, err := c.Run(ctx, "version")
	if err != nil {
		return "", err
	}

	matched := tfVersionRe.FindStringSubmatch(stdout)
	if len(matched) != 2 {
		return "", fmt.Errorf("failed to parse terraform version: %s", stdout)
	}
	version := matched[1]
	return version, nil
}
