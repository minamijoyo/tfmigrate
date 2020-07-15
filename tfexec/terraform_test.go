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

func TestTerraformCLIInit(t *testing.T) {
	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		dir          string
		opts         []string
		ok           bool
	}{
		{
			desc: "no dir and no opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "init"},
					exitCode: 0,
				},
			},
			ok: true,
		},
		{
			desc: "failed to run terraform init",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "init"},
					exitCode: 1,
				},
			},
			ok: false,
		},
		{
			desc: "with dir",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "init", "foo"},
					exitCode: 0,
				},
			},
			dir: "foo",
			ok:  true,
		},
		{
			desc: "with opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "init", "-input=false", "-no-color"},
					exitCode: 0,
				},
			},
			opts: []string{"-input=false", "-no-color"},
			ok:   true,
		},
		{
			desc: "with dir and opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "init", "-input=false", "-no-color", "foo"},
					exitCode: 0,
				},
			},
			dir:  "foo",
			opts: []string{"-input=false", "-no-color"},
			ok:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			err := terraformCLI.Init(context.Background(), tc.dir, tc.opts...)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}
		})
	}
}

func TestTerraformCLIApply(t *testing.T) {
	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		dirOrPlan    string
		opts         []string
		ok           bool
	}{
		{
			desc: "no dir and no opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "apply"},
					exitCode: 0,
				},
			},
			ok: true,
		},
		{
			desc: "failed to run terraform apply",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "apply"},
					exitCode: 1,
				},
			},
			ok: false,
		},
		{
			desc: "with dir",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "apply", "foo"},
					exitCode: 0,
				},
			},
			dirOrPlan: "foo",
			ok:        true,
		},
		{
			desc: "with opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "apply", "-input=false", "-no-color"},
					exitCode: 0,
				},
			},
			opts: []string{"-input=false", "-no-color"},
			ok:   true,
		},
		{
			desc: "with dir and opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "apply", "-input=false", "-no-color", "foo"},
					exitCode: 0,
				},
			},
			dirOrPlan: "foo",
			opts:      []string{"-input=false", "-no-color"},
			ok:        true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			err := terraformCLI.Apply(context.Background(), tc.dirOrPlan, tc.opts...)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
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
