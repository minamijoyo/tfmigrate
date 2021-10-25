package tfexec

import (
	"context"
	"io/ioutil"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestTerraformCLIStateRm(t *testing.T) {
	state := NewState([]byte("dummy state"))
	stateOut := NewState([]byte("dummy state out"))

	// mock writing state to a temporary file.
	runFunc := func(args ...string) error {
		for _, arg := range args {
			// The terraform state rm doesn't have -state-out option.
			// It updates the inpute state in-place.
			if strings.HasPrefix(arg, "-state=") {
				stateFile := arg[len("-state="):]
				return ioutil.WriteFile(stateFile, stateOut.Bytes(), 0644)
			}
		}
		// if the -state option is not set, nothing to do.
		return nil
	}

	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		state        *State
		addresses    []string
		opts         []string
		want         *State
		ok           bool
	}{
		{
			desc: "no opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "rm", "aws_security_group.foo", "aws_security_group.bar"},
					argsRe:   regexp.MustCompile(`^terraform state rm aws_security_group.foo aws_security_group.bar$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:     nil,
			addresses: []string{"aws_security_group.foo", "aws_security_group.bar"},
			want:      nil,
			ok:        true,
		},
		{
			desc: "failed to run terraform state rm",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "rm", "aws_security_group.foo", "aws_security_group.bar"},
					argsRe:   regexp.MustCompile(`^terraform state rm aws_security_group.foo aws_security_group.bar$`),
					runFunc:  runFunc,
					exitCode: 1,
				},
			},
			state:     nil,
			addresses: []string{"aws_security_group.foo", "aws_security_group.bar"},
			want:      nil,
			ok:        false,
		},
		{
			desc: "with opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "rm", "-lock=true", "-lock-timeout=10s", "aws_security_group.foo", "aws_security_group.bar"},
					argsRe:   regexp.MustCompile(`^terraform state rm -lock=true -lock-timeout=10s aws_security_group.foo aws_security_group.bar$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:     nil,
			addresses: []string{"aws_security_group.foo", "aws_security_group.bar"},
			opts:      []string{"-lock=true", "-lock-timeout=10s"},
			want:      nil,
			ok:        true,
		},
		{
			desc: "with state",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "rm", "-state=/path/to/tempfile", "-lock=true", "-lock-timeout=10s", "aws_security_group.foo", "aws_security_group.bar"},
					argsRe:   regexp.MustCompile(`^terraform state rm -state=.+ -lock=true -lock-timeout=10s aws_security_group.foo aws_security_group.bar$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:     state,
			addresses: []string{"aws_security_group.foo", "aws_security_group.bar"},
			opts:      []string{"-lock=true", "-lock-timeout=10s"},
			want:      stateOut,
			ok:        true,
		},
		{
			desc: "with state and -state= (conflict error)",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "rm", "-state=/path/to/tempfile", "-lock=true", "-state=foo.tfstate", "aws_security_group.foo", "aws_security_group.bar"},
					argsRe:   regexp.MustCompile(`^terraform state rm -state=.+ -lock=true -state=foo.tfstate aws_security_group.foo aws_security_group.bar$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:     state,
			addresses: []string{"aws_security_group.foo", "aws_security_group.bar"},
			opts:      []string{"-lock=true", "-state=foo.tfstate"},
			want:      nil,
			ok:        false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			got, err := terraformCLI.StateRm(context.Background(), tc.state, tc.addresses, tc.opts...)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}
			if tc.ok {
				if tc.want != nil {
					if !reflect.DeepEqual(got.Bytes(), tc.want.Bytes()) {
						t.Errorf("got: %v, want: %v", got, tc.want)
					}
				} else { // tc.want == nil
					if got != nil {
						t.Errorf("got: %v, want: %v", got, tc.want)
					}
				}
			}
		})
	}
}

func TestAccTerraformCLIStateRm(t *testing.T) {
	SkipUnlessAcceptanceTestEnabled(t)

	source := `
resource "null_resource" "foo" {}
resource "null_resource" "bar" {}
`
	e := SetupTestAcc(t, source)
	terraformCLI := NewTerraformCLI(e)

	err := terraformCLI.Init(context.Background(), "-input=false", "-no-color")
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

	updatedState, err := terraformCLI.StateRm(context.Background(), state, []string{"null_resource.foo"})
	if err != nil {
		t.Fatalf("failed to run terraform state rm: %s", err)
	}

	got, err := terraformCLI.StateList(context.Background(), updatedState, nil)
	if err != nil {
		t.Fatalf("failed to run terraform state list: %s", err)
	}

	want := []string{"null_resource.bar"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got: %v, want: %v", got, want)
	}
}
