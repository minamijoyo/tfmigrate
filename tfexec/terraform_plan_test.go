package tfexec

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestTerraformCLIPlan(t *testing.T) {
	state := NewState([]byte("dummy state"))

	// mock writing plan to a temporary file.
	plan := NewPlan([]byte("dummy plan"))
	runFunc := func(args ...string) error {
		for _, arg := range args {
			if strings.HasPrefix(arg, "-out=") {
				planFile := arg[len("-out="):]
				return ioutil.WriteFile(planFile, plan.Bytes(), 0600)
			}
		}
		return fmt.Errorf("failed to find -out= option: %v", args)
	}

	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		state        *State
		opts         []string
		want         *Plan
		ok           bool
	}{
		{
			desc: "no opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "plan", "-out=/path/to/planfile"},
					argsRe:   regexp.MustCompile(`^terraform plan -out=.+$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state: nil,
			want:  plan,
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
			want:  NewPlan([]byte{}),
			ok:    false,
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
			want:  plan,
			ok:    true,
		},
		{
			desc: "with state",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "plan", "-state=/path/to/tempfile", "-out=/path/to/planfile", "-input=false", "-no-color"},
					argsRe:   regexp.MustCompile(`^terraform plan -state=.+ -out=.+ -input=false -no-color$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			opts:  []string{"-input=false", "-no-color"},
			state: state,
			want:  plan,
			ok:    true,
		},
		{
			desc: "with state and -state= (conflict error)",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "plan", "-state=/path/to/tempfile", "-out=/path/to/planfile", "-input=false", "-state=foo.tfstate"},
					argsRe:   regexp.MustCompile(`^terraform plan -state=\S+ -out=.+ -input=false -no-color -state=foo.tfstate$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			opts:  []string{"-input=false", "-state=foo.tfstate"},
			state: state,
			want:  nil,
			ok:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			got, err := terraformCLI.Plan(context.Background(), tc.state, tc.opts...)
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

func TestAccTerraformCLIPlan(t *testing.T) {
	SkipUnlessAcceptanceTestEnabled(t)

	source := `resource "null_resource" "foo" {}`
	e := SetupTestAcc(t, source)
	terraformCLI := NewTerraformCLI(e)

	err := terraformCLI.Init(context.Background(), nil, "-input=false", "-no-color")
	if err != nil {
		t.Fatalf("failed to run terraform init: %s", err)
	}

	plan, err := terraformCLI.Plan(context.Background(), nil, "-input=false", "-no-color")
	if err != nil {
		t.Fatalf("failed to run terraform plan: %s", err)
	}

	if plan == nil {
		t.Error("plan success but returns nil")
	}
}

func TestAccTerraformCLIPlanWithOut(t *testing.T) {
	SkipUnlessAcceptanceTestEnabled(t)

	source := `resource "null_resource" "foo" {}`
	e := SetupTestAcc(t, source)
	terraformCLI := NewTerraformCLI(e)

	err := terraformCLI.Init(context.Background(), nil, "-input=false", "-no-color")
	if err != nil {
		t.Fatalf("failed to run terraform init: %s", err)
	}

	planOut := "foo.tfplan"
	plan, err := terraformCLI.Plan(context.Background(), nil, "-input=false", "-no-color", "-out="+planOut)
	if err != nil {
		t.Fatalf("failed to run terraform plan: %s", err)
	}

	if plan == nil {
		t.Error("plan success but returns nil")
	}

	if _, err := os.Stat(filepath.Join(e.Dir(), planOut)); os.IsNotExist(err) {
		t.Errorf("failed to find a plan file: %s, err %s", planOut, err)
	}
}
