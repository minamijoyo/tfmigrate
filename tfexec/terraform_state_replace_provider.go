package tfexec

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-version"
)

const (
	// MinimumTerraformVersionForStateReplaceProvider specifies the minimum
	// supported Terraform version for StateReplaceProvider.
	MinimumTerraformVersionForStateReplaceProvider = "0.13"
)

// StateReplaceProvider replaces providers from source to destination address.
// If a state argument is given, use it for the input state.
// It returns the given state.
func (c *terraformCLI) StateReplaceProvider(ctx context.Context, state *State, source string, destination string, opts ...string) (*State, error) {
	constraints, err := version.NewConstraint(fmt.Sprintf(">= %s", MinimumTerraformVersionForStateReplaceProvider))
	if err != nil {
		return nil, err
	}

	v, err := c.Version(ctx)
	if err != nil {
		return nil, err
	}

	if !constraints.Check(v) {
		return nil, fmt.Errorf("configuration uses Terraform version %s; replace-provider action requires Terraform version %s", v, constraints)
	}

	args := []string{"state", "replace-provider"}

	var tmpState *os.File

	if state != nil {
		if hasPrefixOptions(opts, "-state=") {
			return nil, fmt.Errorf("failed to build options. The state argument (!= nil) and the -state= option cannot be set at the same time: state=%v, opts=%v", state, opts)
		}
		tmpState, err = writeTempFile(state.Bytes())
		defer os.Remove(tmpState.Name())
		if err != nil {
			return nil, err
		}
		args = append(args, "-state="+tmpState.Name())
	}

	args = append(args, opts...)
	args = append(args, source, destination)

	_, _, err = c.Run(ctx, args...)
	if err != nil {
		return nil, err
	}

	// Read updated states
	var updatedState *State

	if state != nil {
		bytes, err := os.ReadFile(tmpState.Name())
		if err != nil {
			return nil, err
		}
		updatedState = NewState(bytes)
	}

	return updatedState, nil
}
