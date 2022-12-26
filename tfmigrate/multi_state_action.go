package tfmigrate

import (
	"context"
	"fmt"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

// MultiStateAction abstracts multi state migration operations.
// It's used for moving resources from one state to another.
type MultiStateAction interface {
	// MultiStateUpdate updates given two states and returns new two states.
	MultiStateUpdate(ctx context.Context, fromTf tfexec.TerraformCLI, toTf tfexec.TerraformCLI, fromState *tfexec.State, toState *tfexec.State) (*tfexec.State, *tfexec.State, error)
}

// NewMultiStateActionFromString is a factory method which returns a new
// MultiStateAction from a given string.
// cmdStr is a plain text for state operation.
// This method is useful to build an action from terraform state command.
// Valid formats are the following.
// "mv <source> <destination>"
// "xmv <source> <destination>"
func NewMultiStateActionFromString(cmdStr string) (MultiStateAction, error) {
	args, err := splitStateAction(cmdStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse action: %s, err: %s", cmdStr, err)
	}

	if len(args) == 0 {
		return nil, fmt.Errorf("multi state action is empty: %s", cmdStr)
	}
	actionType := args[0]

	// switch by action type and parse arguments and build an action.
	var action MultiStateAction
	switch actionType {
	case "mv":
		if len(args) != 3 {
			return nil, fmt.Errorf("multi state mv action is invalid: %s", cmdStr)
		}
		src := args[1]
		dst := args[2]
		action = NewMultiStateMvAction(src, dst)

	case "xmv":
		if len(args) != 3 {
			return nil, fmt.Errorf("multi state xmv action is invalid: %s", cmdStr)
		}
		src := args[1]
		dst := args[2]
		action = NewMultiStateXmvAction(src, dst)

	default:
		return nil, fmt.Errorf("unknown multi state action type: %s", cmdStr)
	}

	return action, nil
}
