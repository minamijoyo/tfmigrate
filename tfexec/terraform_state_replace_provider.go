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

	// AcceptableLegacyStateInitError is the error message returned by `terraform
	// init` when a non-legacy Terraform CLI is used against a legacy Terraform state.
	// When invoking `state replace-provider`, it's necessary to first
	// invoke `terraform init`. However, when using a non-legacy Terraform CLI
	// against a legacy Terraform state, this error is expected.
	AcceptableLegacyStateInitError = "Error: Invalid legacy provider address"
)

// SupportsStateReplaceProvider returns true if the terraform version is greater
// than or equal to 0.13.0 and therefore supports the state replace-provider command.
func (c *terraformCLI) SupportsStateReplaceProvider(ctx context.Context) (bool, version.Constraints, error) {
	constraints, err := version.NewConstraint(fmt.Sprintf(">= %s", MinimumTerraformVersionForStateReplaceProvider))
	if err != nil {
		return false, constraints, err
	}

	v, err := c.Version(ctx)
	if err != nil {
		return false, constraints, err
	}

	ver, err := truncatePreReleaseVersion(v)
	if err != nil {
		return false, constraints, err
	}

	if !constraints.Check(ver) {
		return false, constraints, nil
	}

	return true, constraints, nil
}

// StateReplaceProvider replaces providers from source to destination address.
// If a state argument is given, use it for the input state.
// It returns the given state.
func (c *terraformCLI) StateReplaceProvider(ctx context.Context, state *State, source string, destination string, opts ...string) (*State, error) {
	supports, constraints, err := c.SupportsStateReplaceProvider(ctx)
	if err != nil {
		return nil, err
	}

	if !supports {
		return nil, fmt.Errorf("replace-provider action requires Terraform version %s", constraints)
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
