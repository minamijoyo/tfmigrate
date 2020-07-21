package tfexec

import "context"

// StatePull returns the current tfstate from remote.
func (c *terraformCLI) StatePull(ctx context.Context, opts ...string) (*State, error) {
	args := []string{"state", "pull"}
	// At time of writing, there is no valid option in Terraform v0.12/v0.13.
	// It's a room for future extensions not to break the interface.
	args = append(args, opts...)

	stdout, _, err := c.Run(ctx, args...)
	if err != nil {
		return nil, err
	}

	return NewState([]byte(stdout)), nil
}
