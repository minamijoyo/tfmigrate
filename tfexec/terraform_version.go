package tfexec

import (
	"context"
	"fmt"
	"regexp"
	"strings"

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

// truncatePreReleaseVersion is a helper function that removes
// pre-release information.
// The hashicorp/go-version returns false when comparing pre-releases, for
// example 1.6.0-rc1 >= 0.13. This is counter-intuitive for determining the
// presence or absence of a feature, so remove the pre-release information
// before comparing.
func truncatePreReleaseVersion(v *version.Version) (*version.Version, error) {
	if v.Prerelease() == "" {
		return v, nil
	}

	vs, _, _ := strings.Cut(v.String(), "-")

	ver, err := version.NewVersion(vs)
	if err != nil {
		return nil, err
	}

	return ver, nil
}
