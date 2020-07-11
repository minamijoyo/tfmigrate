package tfexec

import (
	"context"
	"regexp"
	"testing"
)

func TestTerraformCLIVersion(t *testing.T) {
	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		want         string
		ok           bool
	}{
		{
			desc: "parse outputs of terraform version",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "version"},
					stdout:   "Terraform v0.13.0-beta2\n",
					stderr:   "",
					exitCode: 0,
				},
			},
			want: "0.13.0-beta2",
			ok:   true,
		},
		{
			desc: "failed to run terraform version",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "version"},
					stdout:   "",
					stderr:   "",
					exitCode: 1,
				},
			},
			want: "",
			ok:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			got, err := terraformCLI.Version(context.Background())
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

func TestTerraformCLIStatePull(t *testing.T) {
	tfstate := `{
  "version": 4,
  "terraform_version": "0.12.28",
  "serial": 0,
  "lineage": "3d2cf549-8051-c117-aaa7-f93cda2674e8",
  "outputs": {},
  "resources": []
}
`
	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		want         State
		ok           bool
	}{
		{
			desc: "print tfstate to stdout",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "pull"},
					stdout:   tfstate,
					stderr:   "",
					exitCode: 0,
				},
			},
			want: State(tfstate),
			ok:   true,
		},
		{
			desc: "failed to run terraform state pull",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "pull"},
					stdout:   "",
					stderr:   "",
					exitCode: 1,
				},
			},
			want: State(""),
			ok:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			got, err := terraformCLI.StatePull(context.Background())
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

func TestTerraformCLIStatePush(t *testing.T) {
	tfstate := `{
  "version": 4,
  "terraform_version": "0.12.28",
  "serial": 0,
  "lineage": "3d2cf549-8051-c117-aaa7-f93cda2674e8",
  "outputs": {},
  "resources": []
}
`
	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		state        State
		ok           bool
	}{
		{
			desc: "state push from memory",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "push", "/path/to/tempfile"},
					argsRe:   regexp.MustCompile(`^terraform state push .+$`),
					stdout:   "",
					stderr:   "",
					exitCode: 0,
				},
			},
			state: State(tfstate),
			ok:    true,
		},
		{
			desc: "failed to run terraform state push",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "push", "/path/to/tempfile"},
					argsRe:   regexp.MustCompile(`^terraform state push .+$`),
					stdout:   "",
					stderr:   "",
					exitCode: 1,
				},
			},
			state: State(tfstate),
			ok:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			err := terraformCLI.StatePush(context.Background(), tc.state)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error")
			}
		})
	}
}
