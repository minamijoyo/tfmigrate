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
		want         string
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
			want:     "1.6.2",
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
			want:     "1.6.0-alpha3",
			ok:       true,
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
			want:     "",
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
			want:     "0.12.28",
			ok:       true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			terraformCLI.SetExecPath(tc.execPath)
			got, err := terraformCLI.Version(context.Background())
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

func TestAccTerraformCLIVersion(t *testing.T) {
	SkipUnlessAcceptanceTestEnabled(t)

	e := NewExecutor("", os.Environ())
	terraformCLI := NewTerraformCLI(e)
	got, err := terraformCLI.Version(context.Background())
	if err != nil {
		t.Fatalf("failed to run terraform version: %s", err)
	}
	if got.String() == "" {
		t.Error("failed to parse terraform version")
	}
	fmt.Printf("got = %s\n", got)
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
