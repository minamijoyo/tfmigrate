package tfexec

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// StateWriteLocal saves the new state to a local state file.
func (c *terraformCLI) StateWriteLocal(ctx context.Context, state *State) error {
	localStateFile := "terraform.tfstate"
	path := filepath.Join(c.Dir(), localStateFile)
	if err := os.WriteFile(path, []byte(state.Bytes()), 0600); err != nil {
		return fmt.Errorf("failed to write %s: %s", localStateFile, err)
	}

	tmpState, err := writeTempFile(state.Bytes())
	defer os.Remove(tmpState.Name())
	if err != nil {
		return err
	}
	return nil
}
