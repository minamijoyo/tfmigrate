package tfexec

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestTerraformCLIRun(t *testing.T) {
	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		args         []string
		execPath     string
		want         string
		ok           bool
	}{
		{
			desc: "run terraform version",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "version"},
					stdout:   "Terraform v0.12.28\n",
					exitCode: 0,
				},
			},
			args: []string{"version"},
			want: "Terraform v0.12.28\n",
			ok:   true,
		},
		{
			desc: "failed to run terraform version",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "version"},
					exitCode: 1,
				},
			},
			args: []string{"version"},
			want: "",
			ok:   false,
		},
		{
			desc: "with execPath (no space)",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform-0.12.28", "version"},
					stdout:   "Terraform v0.12.28\n",
					exitCode: 0,
				},
			},
			args:     []string{"version"},
			execPath: "terraform-0.12.28",
			want:     "Terraform v0.12.28\n",
			ok:       true,
		},
		{
			desc: "with execPath (spaces)",
			mockCommands: []*mockCommand{
				{
					args:     []string{"direnv", "exec", ".", "terraform", "version"},
					stdout:   "Terraform v0.12.28\n",
					exitCode: 0,
				},
			},
			args:     []string{"version"},
			execPath: "direnv exec . terraform",
			want:     "Terraform v0.12.28\n",
			ok:       true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			terraformCLI.SetExecPath(tc.execPath)
			got, _, err := terraformCLI.Run(context.Background(), tc.args...)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error, got = %s", got)
			}
			if got != tc.want {
				t.Errorf("got: %s, want: %s", got, tc.want)
			}
		})
	}
}

func TestAccTerraformCLIOverrideBackendToLocal(t *testing.T) {
	SkipUnlessAcceptanceTestEnabled(t)

	backend := GetTestAccBackendS3Config(t.Name())
	source := `
resource "aws_security_group" "foo" {}
resource "aws_security_group" "bar" {}
`
	workspace := "work1"
	terraformCLI := SetupTestAccWithApply(t, workspace, backend+source)

	updatedSource := `
resource "aws_security_group" "foo2" {}
resource "aws_security_group" "bar" {}
`
	UpdateTestAccSource(t, terraformCLI, backend+updatedSource)

	changed, err := terraformCLI.PlanHasChange(context.Background(), nil)
	if err != nil {
		t.Fatalf("failed to run PlanHasChange: %s", err)
	}
	if !changed {
		t.Fatalf("expect to have changes")
	}

	state, err := terraformCLI.StatePull(context.Background())
	if err != nil {
		t.Fatalf("failed to run terraform state pull: %s", err)
	}

	filename := "_tfexec_override.tf"
	if _, err := os.Stat(filepath.Join(terraformCLI.Dir(), filename)); err == nil {
		t.Fatalf("an override file already exists: %s", err)
	}

	switchBackToRemotekFunc, err := terraformCLI.OverrideBackendToLocal(context.Background(), filename, workspace, false)
	if err != nil {
		t.Fatalf("failed to run OverrideBackendToLocal: %s", err)
	}

	if _, err := os.Stat(filepath.Join(terraformCLI.Dir(), filename)); os.IsNotExist(err) {
		t.Fatalf("the override file does not exist: %s", err)
	}

	updatedState, _, err := terraformCLI.StateMv(context.Background(), state, nil, "aws_security_group.foo", "aws_security_group.foo2")
	if err != nil {
		t.Fatalf("failed to run terraform state mv: %s", err)
	}

	changed, err = terraformCLI.PlanHasChange(context.Background(), updatedState)
	if err != nil {
		t.Fatalf("failed to run PlanHasChange: %s", err)
	}
	if changed {
		t.Fatalf("expect not to have changes")
	}

	switchBackToRemotekFunc()

	if _, err := os.Stat(filepath.Join(terraformCLI.Dir(), filename)); err == nil {
		t.Fatalf("the override file wasn't removed: %s", err)
	}

	changed, err = terraformCLI.PlanHasChange(context.Background(), nil)
	if err != nil {
		t.Fatalf("failed to run PlanHasChange: %s", err)
	}
	if !changed {
		t.Fatalf("expect to have changes")
	}
}

func TestAccTerraformCLIPlanHasChange(t *testing.T) {
	SkipUnlessAcceptanceTestEnabled(t)

	source := `resource "null_resource" "foo" {}`
	e := SetupTestAcc(t, source)
	terraformCLI := NewTerraformCLI(e)

	err := terraformCLI.Init(context.Background(), "-input=false", "-no-color")
	if err != nil {
		t.Fatalf("failed to run terraform init: %s", err)
	}

	changed, err := terraformCLI.PlanHasChange(context.Background(), nil)
	if err != nil {
		t.Fatalf("failed to run PlanHasChange: %s", err)
	}
	if !changed {
		t.Fatalf("expect not to have changes")
	}

	err = terraformCLI.Apply(context.Background(), nil, "-input=false", "-no-color", "-auto-approve")
	if err != nil {
		t.Fatalf("failed to run terraform apply: %s", err)
	}

	changed, err = terraformCLI.PlanHasChange(context.Background(), nil)
	if err != nil {
		t.Fatalf("failed to run PlanHasChange: %s", err)
	}

	if changed {
		t.Fatalf("expect to have changes")
	}
}

func TestGetOptionValue(t *testing.T) {
	cases := []struct {
		desc   string
		opts   []string
		prefix string
		want   string
	}{
		{
			desc:   "found",
			opts:   []string{"-input=false", "-no-color", "-out=foo.tfplan", "-detailed-exitcode"},
			prefix: "-out=",
			want:   "foo.tfplan",
		},
		{
			desc:   "not found",
			opts:   []string{"-input=false", "-no-color", "-detailed-exitcode"},
			prefix: "-out=",
			want:   "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got := getOptionValue(tc.opts, tc.prefix)
			if got != tc.want {
				t.Errorf("got: %s, want: %s", got, tc.want)
			}
		})
	}
}
