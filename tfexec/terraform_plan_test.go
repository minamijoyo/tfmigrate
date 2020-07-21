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

func TestTerraformCLIPlan(t *testing.T) {
	state := State(`{
  "version": 4,
  "terraform_version": "0.12.28",
  "serial": 0,
  "lineage": "3d2cf549-8051-c117-aaa7-f93cda2674e8",
  "outputs": {},
  "resources": []
}
`)

	// mock writing plan to a temporary file.
	plan := []byte("dummy plan")
	runFunc := func(args ...string) error {
		for _, arg := range args {
			if strings.HasPrefix(arg, "-out=") {
				planFile := arg[len("-out="):]
				return ioutil.WriteFile(planFile, plan, 0644)
			}
		}
		return fmt.Errorf("failed to find -out= option: %v", args)
	}

	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		state        *State
		dir          string
		opts         []string
		want         Plan
		ok           bool
	}{
		{
			desc: "no dir and no opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "plan", "-out=/path/to/planfile"},
					argsRe:   regexp.MustCompile(`^terraform plan -out=.+$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state: nil,
			want:  Plan(plan),
			ok:    true,
		},
		{
			desc: "failed to run terraform plan",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "plan", "-out=/path/to/planfile"},
					argsRe:   regexp.MustCompile(`^terraform plan -out=.+$`),
					exitCode: 1,
				},
			},
			state: nil,
			want:  Plan([]byte{}),
			ok:    false,
		},
		{
			desc: "with dir",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "plan", "-out=/path/to/planfile", "foo"},
					argsRe:   regexp.MustCompile(`^terraform plan -out=.+ foo$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			dir:   "foo",
			state: nil,
			want:  Plan(plan),
			ok:    true,
		},
		{
			desc: "with opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "plan", "-out=/path/to/planfile", "-input=false", "-no-color"},
					argsRe:   regexp.MustCompile(`^terraform plan -out=.+ -input=false -no-color$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			opts:  []string{"-input=false", "-no-color"},
			state: nil,
			want:  Plan(plan),
			ok:    true,
		},
		{
			desc: "with dir and opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "plan", "-out=/path/to/planfile", "-input=false", "-no-color", "foo"},
					argsRe:   regexp.MustCompile(`^terraform plan -out=.+ -input=false -no-color foo$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			dir:   "foo",
			opts:  []string{"-input=false", "-no-color"},
			state: nil,
			want:  Plan(plan),
			ok:    true,
		},
		{
			desc: "with state",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "plan", "-state=/path/to/tempfile", "-out=/path/to/planfile", "-input=false", "-no-color", "foo"},
					argsRe:   regexp.MustCompile(`^terraform plan -state=.+ -out=.+ -input=false -no-color foo$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			dir:   "foo",
			opts:  []string{"-input=false", "-no-color"},
			state: &state,
			want:  Plan(plan),
			ok:    true,
		},
		{
			desc: "with state and -state= (conflict error)",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "plan", "-state=/path/to/tempfile", "-out=/path/to/planfile", "-input=false", "-state=foo.tfstate", "foo"},
					argsRe:   regexp.MustCompile(`^terraform plan -state=\S+ -out=.+ -input=false -no-color -state=foo.tfstate foo$`),
					runFunc:  runFunc,
					exitCode: 1,
				},
			},
			dir:   "foo",
			opts:  []string{"-input=false", "-state=foo.tfstate"},
			state: &state,
			want:  nil,
			ok:    false,
		},
		{
			desc: "with -out= (conflict error)",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "plan", "-state=/path/to/tempfile", "-out=/path/to/planfile", "-input=false", "-out=foo.tfplan", "foo"},
					argsRe:   regexp.MustCompile(`^terraform plan -state=.+ -out=\S+ -input=false -no-color -out=foo.tfplan foo$`),
					runFunc:  runFunc,
					exitCode: 1,
				},
			},
			dir:   "foo",
			opts:  []string{"-input=false", "-out=foo.tfplan"},
			state: &state,
			want:  nil,
			ok:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			got, err := terraformCLI.Plan(context.Background(), tc.state, tc.dir, tc.opts...)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got: %v, want: %v", got, tc.want)
			}
		})
	}
}
