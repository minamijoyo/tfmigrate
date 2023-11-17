package tfexec

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/go-version"
)

func TestTerraformCLIVersion(t *testing.T) {
	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		execPath     string
		execType     string
		version      string
		ok           bool
	}{
		{
			desc: "terraform version",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "version"},
					stdout:   "Terraform v1.6.2\n",
					exitCode: 0,
				},
			},
			execPath: "terraform",
			execType: "terraform",
			version:  "1.6.2",
			ok:       true,
		},
		{
			desc: "tofu version",
			mockCommands: []*mockCommand{
				{
					args:     []string{"tofu", "version"},
					stdout:   "OpenTofu v1.6.0-alpha3\n",
					exitCode: 0,
				},
			},
			execPath: "tofu",
			execType: "opentofu",
			version:  "1.6.0-alpha3",
			ok:       true,
		},
		{
			desc: "with wrapper",
			mockCommands: []*mockCommand{
				{
					args:     []string{"direnv", "exec", ".", "terraform", "version"},
					stdout:   "Terraform v1.6.2\n",
					exitCode: 0,
				},
			},
			execPath: "direnv exec . terraform",
			execType: "terraform",
			version:  "1.6.2",
			ok:       true,
		},
		{
			desc: "unknown execType",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "version"},
					stdout:   "MyTerraform v1.6.2\n",
					exitCode: 0,
				},
			},
			execPath: "terraform",
			execType: "",
			version:  "",
			ok:       false,
		},
		{
			desc: "failed to run terraform version",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "version"},
					exitCode: 1,
				},
			},
			execPath: "terraform",
			execType: "",
			version:  "",
			ok:       false,
		},
		{
			desc: "with check point warning",
			mockCommands: []*mockCommand{
				{
					args: []string{"terraform", "version"},
					stdout: `Terraform v0.12.28

Your version of Terraform is out of date! The latest version
is 0.12.29. You can update by downloading from https://www.terraform.io/downloads.html
`,
					exitCode: 0,
				},
			},
			execPath: "terraform",
			execType: "terraform",
			version:  "0.12.28",
			ok:       true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			terraformCLI.SetExecPath(tc.execPath)
			execType, version, err := terraformCLI.Version(context.Background())
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error, execType = %s, version = %s", execType, version.String())
			}
			if tc.ok && execType != tc.execType {
				t.Errorf("unexpected execType, got: %s, want: %s", execType, tc.execType)
			}
			if tc.ok && version.String() != tc.version {
				t.Errorf("unexpected version, got: %s, want: %s", version.String(), tc.version)
			}
		})
	}
}

func TestAccTerraformCLIVersion(t *testing.T) {
	SkipUnlessAcceptanceTestEnabled(t)

	e := NewExecutor("", os.Environ())
	terraformCLI := NewTerraformCLI(e)
	execType, version, err := terraformCLI.Version(context.Background())
	if err != nil {
		t.Fatalf("failed to run terraform version: %s", err)
	}
	if execType == "" {
		t.Error("failed to parse terraform execType")
	}
	if version.String() == "" {
		t.Error("failed to parse terraform version")
	}
	fmt.Printf("got: execType = %s, version = %s\n", execType, version)
}

func TestTruncatePreReleaseVersion(t *testing.T) {
	cases := []struct {
		desc string
		v    string
		want string
		ok   bool
	}{
		{
			desc: "pre-release",
			v:    "1.6.0-rc1",
			want: "1.6.0",
			ok:   true,
		},
		{
			desc: "not pre-release",
			v:    "1.6.0",
			want: "1.6.0",
			ok:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			v, err := version.NewVersion(tc.v)
			if err != nil {
				t.Fatalf("failed to parse version: %s", err)
			}
			got, err := truncatePreReleaseVersion(v)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error, got = %s", got)
			}
			if tc.ok && got.String() != tc.want {
				t.Errorf("got: %s, want: %s", got, tc.want)
			}
		})
	}
}
