package tfexec

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/go-version"
)

// tfVersionRe is a pattern to parse outputs from terraform version.
var tfVersionRe = regexp.MustCompile(`^Terraform v(.+)\s*\n`)

// Version returns a version number of Terraform.
func (c *terraformCLI) Version(ctx context.Context) (*version.Version, error) {
	stdout, _, err := c.Run(ctx, "version")
	if err != nil {
		return nil, err
	}

	matched := tfVersionRe.FindStringSubmatch(stdout)
	if len(matched) != 2 {
		return nil, fmt.Errorf("failed to parse terraform version: %s", stdout)
	}
	version, err := version.NewVersion(matched[1])
	if err != nil {
		return nil, err
	}

	return version, nil
}
