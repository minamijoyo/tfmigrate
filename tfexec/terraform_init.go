package tfexec

import (
	"context"
	"fmt"
)

// Init initializes the current work directory.
func (c *terraformCLI) Init(ctx context.Context, backendConfig []string, opts ...string) error {
	args := []string{"init"}
	args = append(args, opts...)
	for _, v := range backendConfig {
		args = append(args, fmt.Sprintf(`-backend-config=%s`, v))
	}
	_, _, err := c.Run(ctx, args...)
	return err
}
