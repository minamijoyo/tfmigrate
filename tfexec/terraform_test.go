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

func TestAccTerraformCLIOverrideBackendToRemote(t *testing.T) {
	SkipUnlessAcceptanceTestEnabled(t)

	backend := GetTestAccBackendS3Config()
	source := `
resource "aws_security_group" "foo" {}
resource "aws_security_group" "bar" {}
resource "aws_security_group" "baz" {}
`

	tf := SetupTestAccWithApply(t, backend+source)

	filename := "_tfexec_override.tf"
	if _, err := os.Stat(filepath.Join(tf.Dir(), filename)); err == nil {
		t.Fatalf("an override file already exists: %s", err)
	}

	switchBackToRemotekFunc, err := tf.OverrideBackendToRemote(context.Background(), filename)
	if err != nil {
		t.Fatalf("failed to run OverrideBackendToRemote: %s", err)
	}

	if _, err := os.Stat(filepath.Join(tf.Dir(), filename)); os.IsNotExist(err) {
		t.Fatalf("the override file does not exist: %s", err)
	}

	switchBackToRemotekFunc()

	if _, err := os.Stat(filepath.Join(tf.Dir(), filename)); err == nil {
		t.Fatalf("the override file wasn't removed: %s", err)
	}
}

func TestAccTerraformCLIPlanHasChange(t *testing.T) {
	SkipUnlessAcceptanceTestEnabled(t)

	source := `resource "null_resource" "foo" {}`
	e := SetupTestAcc(t, source)
	terraformCLI := NewTerraformCLI(e)

	err := terraformCLI.Init(context.Background(), "", "-input=false", "-no-color")
	if err != nil {
		t.Fatalf("failed to run terraform init: %s", err)
	}

	changed, err := terraformCLI.PlanHasChange(context.Background(), nil, "")
	if err != nil {
		t.Fatalf("failed to run PlanHasChange: %s", err)
	}
	if !changed {
		t.Fatalf("expect not to have changes")
	}

	err = terraformCLI.Apply(context.Background(), nil, "", "-input=false", "-no-color", "-auto-approve")
	if err != nil {
		t.Fatalf("failed to run terraform apply: %s", err)
	}

	changed, err = terraformCLI.PlanHasChange(context.Background(), nil, "")
	if err != nil {
		t.Fatalf("failed to run PlanHasChange: %s", err)
	}

	if changed {
		t.Fatalf("expect to have changes")
	}
}
