package tfexec

import (
	"context"
	"fmt"
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
resource "null_resource" "foo" {}
resource "null_resource" "bar" {}
`
	workspace := "work1"
	terraformCLI := SetupTestAccWithApply(t, workspace, backend+source)

	updatedSource := `
resource "null_resource" "foo2" {}
resource "null_resource" "bar" {}
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

	switchBackToRemotekFunc, err := terraformCLI.OverrideBackendToLocal(context.Background(), filename, workspace, false, nil)
	if err != nil {
		t.Fatalf("failed to run OverrideBackendToLocal: %s", err)
	}

	if _, err := os.Stat(filepath.Join(terraformCLI.Dir(), filename)); os.IsNotExist(err) {
		t.Fatalf("the override file does not exist: %s", err)
	}

	updatedState, _, err := terraformCLI.StateMv(context.Background(), state, nil, "null_resource.foo", "null_resource.foo2")
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

func TestAccTerraformCLIOverrideBackendToLocalWithBackendConfig(t *testing.T) {
	SkipUnlessAcceptanceTestEnabled(t)

	endpoint := "http://localhost:4566"
	localstackEndpoint := os.Getenv("LOCALSTACK_ENDPOINT")
	if len(localstackEndpoint) > 0 {
		endpoint = localstackEndpoint
	}

	backend := fmt.Sprintf(`
terraform {
  # https://www.terraform.io/docs/backends/types/s3.html
  backend "s3" {
    region = "ap-northeast-1"
    // bucket = "tfstate-test"
    key    = "%s/terraform.tfstate"

    // mock s3 endpoint with localstack
    endpoint                    = "%s"
    access_key                  = "dummy"
    secret_key                  = "dummy"
    skip_credentials_validation = true
    skip_metadata_api_check     = true
    force_path_style            = true
  }
}
`, t.Name(), endpoint)
	source := `
resource "null_resource" "foo" {}
resource "null_resource" "bar" {}
`
	workspace := "work1"
	backendConfig := []string{"bucket=tfstate-test"}
	e := SetupTestAcc(t, source+backend)
	terraformCLI := NewTerraformCLI(e)
	ctx := context.Background()

	var args = []string{"-input=false", "-no-color"}
	for _, b := range backendConfig {
		args = append(args, fmt.Sprintf("-backend-config=%s", b))
	}
	err := terraformCLI.Init(ctx, args...)
	if err != nil {
		t.Fatalf("failed to run terraform init: %s", err)
	}

	//default workspace always exists so don't try to create it
	if workspace != "default" {
		err = terraformCLI.WorkspaceNew(ctx, workspace)
		if err != nil {
			t.Fatalf("failed to run terraform workspace new %s : %s", workspace, err)
		}
	}

	err = terraformCLI.Apply(ctx, nil, "-input=false", "-no-color", "-auto-approve")
	if err != nil {
		t.Fatalf("failed to run terraform apply: %s", err)
	}

	// destroy resources after each test not to have any state.
	t.Cleanup(func() {
		err := terraformCLI.Destroy(ctx, "-input=false", "-no-color", "-auto-approve")
		if err != nil {
			t.Fatalf("failed to run terraform destroy: %s", err)
		}
	})

	updatedSource := `
resource "null_resource" "foo2" {}
resource "null_resource" "bar" {}
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

	switchBackToRemotekFunc, err := terraformCLI.OverrideBackendToLocal(context.Background(), filename, workspace, false, backendConfig)
	if err != nil {
		t.Fatalf("failed to run OverrideBackendToLocal: %s", err)
	}

	if _, err := os.Stat(filepath.Join(terraformCLI.Dir(), filename)); os.IsNotExist(err) {
		t.Fatalf("the override file does not exist: %s", err)
	}

	updatedState, _, err := terraformCLI.StateMv(context.Background(), state, nil, "null_resource.foo", "null_resource.foo2")
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
