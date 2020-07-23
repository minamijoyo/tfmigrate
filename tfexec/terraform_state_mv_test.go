package tfexec

import (
	"context"
	"fmt"
	"io/ioutil"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestTerraformCLIStateMv(t *testing.T) {
	state := NewState([]byte("dummy state"))
	stateOut := NewState([]byte("dummy state out"))

	// mock writing state to a temporary file.
	runFunc := func(args ...string) error {
		for _, arg := range args {
			if strings.HasPrefix(arg, "-state-out=") {
				stateOutFile := arg[len("-state-out="):]
				return ioutil.WriteFile(stateOutFile, stateOut.Bytes(), 0644)
			}
		}
		return fmt.Errorf("failed to find -state-out= option: %v", args)
	}

	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		state        *State
		source       string
		destination  string
		opts         []string
		want         *State
		ok           bool
	}{
		{
			desc: "no opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "mv", "-state-out=/path/to/out.tfstate", "aws_security_group.foo", "aws_security_group.baz"},
					argsRe:   regexp.MustCompile(`^terraform state mv -state-out=.+ aws_security_group.foo aws_security_group.baz$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:       nil,
			source:      "aws_security_group.foo",
			destination: "aws_security_group.baz",
			want:        stateOut,
			ok:          true,
		},
		{
			desc: "failed to run terraform state mv",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "mv", "-state-out=/path/to/out.tfstate", "aws_security_group.foo", "aws_security_group.baz"},
					argsRe:   regexp.MustCompile(`^terraform state mv -state-out=.+ aws_security_group.foo aws_security_group.baz$`),
					runFunc:  runFunc,
					exitCode: 1,
				},
			},
			state:       nil,
			source:      "aws_security_group.foo",
			destination: "aws_security_group.baz",
			want:        nil,
			ok:          false,
		},
		{
			desc: "with opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "mv", "-state-out=/path/to/out.tfstate", "-lock=true", "-lock-timeout=10s", "aws_security_group.foo", "aws_security_group.baz"},
					argsRe:   regexp.MustCompile(`^terraform state mv -state-out=.+ -lock=true -lock-timeout=10s aws_security_group.foo aws_security_group.baz$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:       nil,
			source:      "aws_security_group.foo",
			destination: "aws_security_group.baz",
			opts:        []string{"-lock=true", "-lock-timeout=10s"},
			want:        stateOut,
			ok:          true,
		},
		{
			desc: "with state",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "mv", "-state=/path/to/tempfile", "-state-out=/path/to/out.tfstate", "-lock=true", "-lock-timeout=10s", "aws_security_group.foo", "aws_security_group.baz"},
					argsRe:   regexp.MustCompile(`^terraform state mv -state=.+ -state-out=.+ -lock=true -lock-timeout=10s aws_security_group.foo aws_security_group.baz$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:       state,
			source:      "aws_security_group.foo",
			destination: "aws_security_group.baz",
			opts:        []string{"-lock=true", "-lock-timeout=10s"},
			want:        stateOut,
			ok:          true,
		},
		{
			desc: "with state and -state= (conflict error)",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "mv", "-state=/path/to/tempfile", "-state-out=/path/to/out.tfstate", "-lock=true", "-state=foo.tfstate", "aws_security_group.foo", "aws_security_group.baz"},
					argsRe:   regexp.MustCompile(`^terraform state mv -state=.+ -state-out=.+ -lock=true -state=foo.tfstate aws_security_group.foo aws_security_group.baz$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:       state,
			source:      "aws_security_group.foo",
			destination: "aws_security_group.baz",
			opts:        []string{"-lock=true", "-state=foo.tfstate"},
			want:        nil,
			ok:          false,
		},
		{
			desc: "with -state-out= (conflict error)",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "mv", "-state=/path/to/tempfile", "-state-out=/path/to/out.tfstate", "-lock=true", "-state-out=foo.tfstate", "aws_security_group.foo", "aws_security_group.baz"},
					argsRe:   regexp.MustCompile(`^terraform state mv -state=.+ -state-out=.+ -lock=true -state-out=foo.tfstate aws_security_group.foo aws_security_group.baz$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:       state,
			source:      "aws_security_group.foo",
			destination: "aws_security_group.baz",
			opts:        []string{"-lock=true", "-state-out=foo.tfstate"},
			want:        nil,
			ok:          false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			got, err := terraformCLI.StateMv(context.Background(), tc.state, tc.source, tc.destination, tc.opts...)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}
			if tc.ok && !reflect.DeepEqual(got.Bytes(), tc.want.Bytes()) {
				t.Errorf("got: %v, want: %v", got, tc.want)
			}
		})
	}
}
