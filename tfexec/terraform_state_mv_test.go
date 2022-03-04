package tfexec

import (
	"context"
	"io/ioutil"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestTerraformCLIStateMv(t *testing.T) {
	state := NewState([]byte("dummy state"))
	stateOut := NewState([]byte("dummy state out"))
	updatedState := NewState([]byte("updated dummy state"))
	updatedStateOut := NewState([]byte("updated dummy state out"))

	// mock writing state to a temporary file.
	runFunc := func(args ...string) error {
		for _, arg := range args {
			if strings.HasPrefix(arg, "-state=") {
				stateFile := arg[len("-state="):]
				err := ioutil.WriteFile(stateFile, updatedState.Bytes(), 0600)
				if err != nil {
					return err
				}
			}
			if strings.HasPrefix(arg, "-state-out=") {
				stateOutFile := arg[len("-state-out="):]
				err := ioutil.WriteFile(stateOutFile, updatedStateOut.Bytes(), 0600)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	cases := []struct {
		desc            string
		mockCommands    []*mockCommand
		state           *State
		stateOut        *State
		source          string
		destination     string
		opts            []string
		updatedState    *State
		updatedStateOut *State
		ok              bool
	}{
		{
			desc: "no opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "mv", "aws_security_group.foo", "aws_security_group.baz"},
					argsRe:   regexp.MustCompile(`^terraform state mv aws_security_group.foo aws_security_group.baz$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:           nil,
			stateOut:        nil,
			source:          "aws_security_group.foo",
			destination:     "aws_security_group.baz",
			updatedState:    nil,
			updatedStateOut: nil,
			ok:              true,
		},
		{
			desc: "failed to run terraform state mv",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "mv", "aws_security_group.foo", "aws_security_group.baz"},
					argsRe:   regexp.MustCompile(`^terraform state mv aws_security_group.foo aws_security_group.baz$`),
					runFunc:  runFunc,
					exitCode: 1,
				},
			},
			state:           nil,
			stateOut:        nil,
			source:          "aws_security_group.foo",
			destination:     "aws_security_group.baz",
			updatedState:    nil,
			updatedStateOut: nil,
			ok:              false,
		},
		{
			desc: "with opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "mv", "-lock=true", "-lock-timeout=10s", "aws_security_group.foo", "aws_security_group.baz"},
					argsRe:   regexp.MustCompile(`^terraform state mv -lock=true -lock-timeout=10s aws_security_group.foo aws_security_group.baz$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:           nil,
			stateOut:        nil,
			source:          "aws_security_group.foo",
			destination:     "aws_security_group.baz",
			opts:            []string{"-lock=true", "-lock-timeout=10s"},
			updatedState:    nil,
			updatedStateOut: nil,
			ok:              true,
		},
		{
			desc: "with state",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "mv", "-state=/path/to/tempfile", "-lock=true", "-lock-timeout=10s", "aws_security_group.foo", "aws_security_group.baz"},
					argsRe:   regexp.MustCompile(`^terraform state mv -state=.+ -lock=true -lock-timeout=10s aws_security_group.foo aws_security_group.baz$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:           state,
			stateOut:        nil,
			source:          "aws_security_group.foo",
			destination:     "aws_security_group.baz",
			opts:            []string{"-lock=true", "-lock-timeout=10s"},
			updatedState:    updatedState,
			updatedStateOut: nil,
			ok:              true,
		},
		{
			desc: "with state and -state= (conflict error)",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "mv", "-state=/path/to/tempfile", "-lock=true", "-state=foo.tfstate", "aws_security_group.foo", "aws_security_group.baz"},
					argsRe:   regexp.MustCompile(`^terraform state mv -state=.+ -lock=true -state=foo.tfstate aws_security_group.foo aws_security_group.baz$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:           state,
			stateOut:        nil,
			source:          "aws_security_group.foo",
			destination:     "aws_security_group.baz",
			opts:            []string{"-lock=true", "-state=foo.tfstate"},
			updatedState:    nil,
			updatedStateOut: nil,
			ok:              false,
		},
		{
			desc: "with stateOut",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "mv", "-state-out=/path/to/tempfile", "-lock=true", "-lock-timeout=10s", "aws_security_group.foo", "aws_security_group.baz"},
					argsRe:   regexp.MustCompile(`^terraform state mv -state-out=.+ -lock=true -lock-timeout=10s aws_security_group.foo aws_security_group.baz$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:           nil,
			stateOut:        stateOut,
			source:          "aws_security_group.foo",
			destination:     "aws_security_group.baz",
			opts:            []string{"-lock=true", "-lock-timeout=10s"},
			updatedState:    nil,
			updatedStateOut: updatedStateOut,
			ok:              true,
		},
		{
			desc: "with -state-out= (conflict error)",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "mv", "-state-out=/path/to/out.tfstate", "-lock=true", "-state-out=foo.tfstate", "aws_security_group.foo", "aws_security_group.baz"},
					argsRe:   regexp.MustCompile(`^terraform state mv -state-out=.+ -lock=true -state-out=foo.tfstate aws_security_group.foo aws_security_group.baz$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:           nil,
			stateOut:        stateOut,
			source:          "aws_security_group.foo",
			destination:     "aws_security_group.baz",
			opts:            []string{"-lock=true", "-state-out=foo.tfstate"},
			updatedState:    nil,
			updatedStateOut: nil,
			ok:              false,
		},
		{
			desc: "with state and stateOut",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "mv", "-state=/path/to/stateTempfile", "-state-out=/path/to/stateOutTempfile", "-lock=true", "-lock-timeout=10s", "aws_security_group.foo", "aws_security_group.baz"},
					argsRe:   regexp.MustCompile(`^terraform state mv -state=.+ -state-out=.+ -lock=true -lock-timeout=10s aws_security_group.foo aws_security_group.baz$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:           state,
			stateOut:        stateOut,
			source:          "aws_security_group.foo",
			destination:     "aws_security_group.baz",
			opts:            []string{"-lock=true", "-lock-timeout=10s"},
			updatedState:    updatedState,
			updatedStateOut: updatedStateOut,
			ok:              true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			gotState, gotStateOut, err := terraformCLI.StateMv(context.Background(), tc.state, tc.stateOut, tc.source, tc.destination, tc.opts...)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}
			if tc.ok {
				if tc.updatedState != nil {
					if !reflect.DeepEqual(gotState.Bytes(), tc.updatedState.Bytes()) {
						t.Errorf("got state: %v, want state: %v", gotState, tc.updatedState)
					}
				} else { // tc.updateState == nil
					if gotState != nil {
						t.Errorf("got state: %v, want state: %v", gotState, tc.updatedState)
					}
				}
				if tc.updatedStateOut != nil {
					if !reflect.DeepEqual(gotStateOut.Bytes(), tc.updatedStateOut.Bytes()) {
						t.Errorf("got stateOut: %v, want stateOut: %v", gotStateOut, tc.updatedStateOut)
					}
				} else { // tc.updateStateOut == nil
					if gotStateOut != nil {
						t.Errorf("got stateOut: %v, want stateOut: %v", gotStateOut, tc.updatedStateOut)
					}
				}
			}
		})
	}
}

func TestAccTerraformCLIStateMv(t *testing.T) {
	SkipUnlessAcceptanceTestEnabled(t)

	source := `
resource "null_resource" "foo" {}
resource "null_resource" "bar" {}
`
	e := SetupTestAcc(t, source)
	terraformCLI := NewTerraformCLI(e)

	_, err := terraformCLI.Init(context.Background(), "-input=false", "-no-color")
	if err != nil {
		t.Fatalf("failed to run terraform init: %s", err)
	}

	err = terraformCLI.Apply(context.Background(), nil, "-input=false", "-no-color", "-auto-approve")
	if err != nil {
		t.Fatalf("failed to run terraform apply: %s", err)
	}

	state, err := terraformCLI.StatePull(context.Background())
	if err != nil {
		t.Fatalf("failed to run terraform state pull: %s", err)
	}

	updatedState, _, err := terraformCLI.StateMv(context.Background(), state, nil, "null_resource.foo", "null_resource.baz")
	if err != nil {
		t.Fatalf("failed to run terraform state mv: %s", err)
	}

	got, err := terraformCLI.StateList(context.Background(), updatedState, nil)
	if err != nil {
		t.Fatalf("failed to run terraform state list: %s", err)
	}

	want := []string{"null_resource.baz", "null_resource.bar"}
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got: %v, want: %v", got, want)
	}
}

func TestAccTerraformCLIStateMvWithStateOut(t *testing.T) {
	SkipUnlessAcceptanceTestEnabled(t)

	source := `
resource "null_resource" "foo" {}
resource "null_resource" "bar" {}
`
	e := SetupTestAcc(t, source)
	terraformCLI := NewTerraformCLI(e)

	_, err := terraformCLI.Init(context.Background(), "-input=false", "-no-color")
	if err != nil {
		t.Fatalf("failed to run terraform init: %s", err)
	}

	err = terraformCLI.Apply(context.Background(), nil, "-input=false", "-no-color", "-auto-approve")
	if err != nil {
		t.Fatalf("failed to run terraform apply: %s", err)
	}

	state, err := terraformCLI.StatePull(context.Background())
	if err != nil {
		t.Fatalf("failed to run terraform state pull: %s", err)
	}

	stateOut := NewState([]byte{})
	updatedState, updatedStateOut, err := terraformCLI.StateMv(context.Background(), state, stateOut, "null_resource.foo", "null_resource.baz")
	if err != nil {
		t.Fatalf("failed to run terraform state mv: %s", err)
	}

	gotState, err := terraformCLI.StateList(context.Background(), updatedState, nil)
	if err != nil {
		t.Fatalf("failed to run terraform state list: %s", err)
	}

	wantState := []string{"null_resource.bar"}
	if !reflect.DeepEqual(gotState, wantState) {
		t.Errorf("got state: %v, want state: %v", gotState, wantState)
	}

	gotStateOut, err := terraformCLI.StateList(context.Background(), updatedStateOut, nil)
	if err != nil {
		t.Fatalf("failed to run terraform state list: %s", err)
	}

	wantStateOut := []string{"null_resource.baz"}
	if !reflect.DeepEqual(gotStateOut, wantStateOut) {
		t.Errorf("got stateOut: %v, want stateOut: %v", gotStateOut, wantStateOut)
	}
}
