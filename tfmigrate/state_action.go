package tfmigrate

import (
	"context"
	"fmt"
	"strings"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

// StateAction abstracts state migration operations.
type StateAction interface {
	// StateUpdate updates a given state and returns a new state.
	StateUpdate(ctx context.Context, tf tfexec.TerraformCLI, state *tfexec.State) (*tfexec.State, error)
}

// NewStateActionFromString is a factory method which returns a new StateAction
// from a given string.
// cmdStr is a plain text for state operation.
// This method is useful to build an action from terraform state command.
// Valid formats are the following.
// "mv <source> <destination>"
// "rm <addresses>...
// "import <address> <id>"
func NewStateActionFromString(cmdStr string) (StateAction, error) {
	// split cmdStr using Fields instead of Split to allow cmdStr to have duplicated white spaces.
	args := strings.Fields(cmdStr)
	if len(args) == 0 {
		return nil, fmt.Errorf("state action is empty: %s", cmdStr)
	}
	actionType := args[0]

	// switch by action type and parse arguments and build an action.
	var action StateAction
	switch actionType {
	case "mv":
		if len(args) != 3 {
			return nil, fmt.Errorf("state mv action is invalid: %s", cmdStr)
		}
		src := args[1]
		dst := args[2]
		action = NewStateMvAction(src, dst)

	case "rm":
		if len(args) < 2 {
			return nil, fmt.Errorf("state rm action is invalid: %s", cmdStr)
		}
		addrs := args[1:]
		action = NewStateRmAction(addrs)

	case "import":
		if len(args) != 3 {
			return nil, fmt.Errorf("state import action is invalid: %s", cmdStr)
		}
		addr := args[1]
		id := args[2]
		action = NewStateImportAction(addr, id)

	default:
		return nil, fmt.Errorf("unknown state action type: %s", cmdStr)
	}

	return action, nil
}
