package tfexec

import (
	"context"
	"fmt"
	"io/ioutil"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestTerraformCLIImport(t *testing.T) {
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
		address      string
		id           string
		opts         []string
		want         *State
		ok           bool
	}{
		{
			desc: "no opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "import", "-state-out=/path/to/out.tfstate", "aws_security_group.foo", "sg-12345678"},
					argsRe:   regexp.MustCompile(`^terraform import -state-out=.+ aws_security_group.foo sg-12345678$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:   nil,
			address: "aws_security_group.foo",
			id:      "sg-12345678",
			want:    stateOut,
			ok:      true,
		},
		{
			desc: "failed to run terraform import",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "import", "-state-out=/path/to/out.tfstate", "aws_security_group.foo", "sg-12345678"},
					argsRe:   regexp.MustCompile(`^terraform import -state-out=.+ aws_security_group.foo sg-12345678$`),
					runFunc:  runFunc,
					exitCode: 1,
				},
			},
			state:   nil,
			address: "aws_security_group.foo",
			id:      "sg-12345678",
			want:    nil,
			ok:      false,
		},
		{
			desc: "with opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "import", "-state-out=/path/to/out.tfstate", "-input=false", "-no-color", "aws_security_group.foo", "sg-12345678"},
					argsRe:   regexp.MustCompile(`^terraform import -state-out=.+ -input=false -no-color aws_security_group.foo sg-12345678$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:   nil,
			address: "aws_security_group.foo",
			id:      "sg-12345678",
			opts:    []string{"-input=false", "-no-color"},
			want:    stateOut,
			ok:      true,
		},
		{
			desc: "with state",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "import", "-state=/path/to/tempfile", "-state-out=/path/to/out.tfstate", "-input=false", "-no-color", "aws_security_group.foo", "sg-12345678"},
					argsRe:   regexp.MustCompile(`^terraform import -state=.+ -state-out=.+ -input=false -no-color aws_security_group.foo sg-12345678$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:   state,
			address: "aws_security_group.foo",
			id:      "sg-12345678",
			opts:    []string{"-input=false", "-no-color"},
			want:    stateOut,
			ok:      true,
		},
		{
			desc: "with state and -state= (conflict error)",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "import", "-state=/path/to/tempfile", "-state-out=/path/to/out.tfstate", "-input=false", "-state=foo.tfstate", "aws_security_group.foo", "sg-12345678"},
					argsRe:   regexp.MustCompile(`^terraform import -state=.+ -state-out=.+ -input=false -state=foo.tfstate aws_security_group.foo sg-12345678$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:   state,
			address: "aws_security_group.foo",
			id:      "sg-12345678",
			opts:    []string{"-input=false", "-state=foo.tfstate"},
			want:    nil,
			ok:      false,
		},
		{
			desc: "with -state-out= (conflict error)",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "import", "-state=/path/to/tempfile", "-state-out=/path/to/out.tfstate", "-input=false", "-state-out=foo.tfstate", "aws_security_group.foo", "sg-12345678"},
					argsRe:   regexp.MustCompile(`^terraform import -state=.+ -state-out=.+ -input=false -state-out=foo.tfstate aws_security_group.foo sg-12345678$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:   state,
			address: "aws_security_group.foo",
			id:      "sg-12345678",
			opts:    []string{"-input=false", "-state-out=foo.tfstate"},
			want:    nil,
			ok:      false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			got, err := terraformCLI.Import(context.Background(), tc.state, tc.address, tc.id, tc.opts...)
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

func TestAccTerraformCLIImport(t *testing.T) {
	SkipUnlessAcceptanceTestEnabled(t)

	source := `resource "time_static" "foo" {}`
	e := SetupTestAcc(t, source)
	terraformCLI := NewTerraformCLI(e)

	err := terraformCLI.Init(context.Background(), "-input=false", "-no-color")
	if err != nil {
		t.Fatalf("failed to run terraform init: %s", err)
	}

	_, err = terraformCLI.Plan(context.Background(), nil, "", "-input=false", "-no-color", "-detailed-exitcode")
	if err != nil {
		if exitErr, ok := err.(*exitError); ok {
			if exitErr.ExitCode() != 2 {
				t.Fatalf("failed to run terraform plan before import (expected diff): %s", err)
			}
		} else {
			t.Fatalf("failed to run terraform plan before import (unexpected error): %s", err)
		}
	}

	// importing a time_static resource accepts any timestamp with format RFC3339.
	now := time.Now().UTC().Format(time.RFC3339)
	state, err := terraformCLI.Import(context.Background(), nil, "time_static.foo", now, "-input=false", "-no-color")
	if err != nil {
		t.Fatalf("failed to run terraform import: %s", err)
	}

	got, err := terraformCLI.StateList(context.Background(), state, nil)
	if err != nil {
		t.Fatalf("failed to run terraform state list: %s", err)
	}

	want := []string{"time_static.foo"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got: %v, want: %v", got, want)
	}

	_, err = terraformCLI.Plan(context.Background(), state, "", "-input=false", "-no-color", "-detailed-exitcode")
	if err != nil {
		t.Fatalf("failed to run terraform plan after import (expected no diff): %s", err)
	}
}
