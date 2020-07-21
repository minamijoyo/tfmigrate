package tfexec

import (
	"context"
	"os"
)

// StatePush pushs a given State to remote.
func (c *terraformCLI) StatePush(ctx context.Context, state State) error {
	tmpState, err := writeTempFile([]byte(state))
	defer os.Remove(tmpState.Name())
	if err != nil {
		return err
	}

	_, _, err = c.Run(ctx, "state", "push", tmpState.Name())
	return err
}
