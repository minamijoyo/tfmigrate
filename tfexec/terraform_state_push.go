package tfexec

import (
	"context"
	"os"
)

// StatePush pushes a given State to remote.
func (c *terraformCLI) StatePush(ctx context.Context, state *State, opts ...string) error {
	args := []string{"state", "push"}
	args = append(args, opts...)

	tmpState, err := writeTempFile(state.Bytes())
	defer os.Remove(tmpState.Name())
	if err != nil {
		return err
	}

	args = append(args, tmpState.Name())
	_, _, err = c.Run(ctx, args...)
	return err
}
