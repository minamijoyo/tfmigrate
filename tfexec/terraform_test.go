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

func TestTerraformCLIRun(t *testing.T) {
	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		args         []string
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
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
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
					stdout:   "Terraform v0.12.28\n",
					exitCode: 0,
				},
			},
			want: "0.12.28",
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

func TestTerraformCLIDestroy(t *testing.T) {
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
					args:     []string{"terraform", "destroy"},
					exitCode: 0,
				},
			},
			ok: true,
		},
		{
			desc: "failed to run terraform destroy",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "destroy"},
					exitCode: 1,
				},
			},
			ok: false,
		},
		{
			desc: "with dir",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "destroy", "foo"},
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
					args:     []string{"terraform", "destroy", "-input=false", "-no-color"},
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
					args:     []string{"terraform", "destroy", "-input=false", "-no-color", "foo"},
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
			err := terraformCLI.Destroy(context.Background(), tc.dir, tc.opts...)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}
		})
	}
}

func TestTerraformCLIStateList(t *testing.T) {
	state := State(testStateListState)
	stdout := `aws_security_group.bar
aws_security_group.foo
`

	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		state        *State
		addresses    []string
		opts         []string
		want         []string
		ok           bool
	}{
		{
			desc: "no addresses and no opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "list"},
					stdout:   stdout,
					exitCode: 0,
				},
			},
			state: nil,
			want:  []string{"aws_security_group.bar", "aws_security_group.foo"},
			ok:    true,
		},
		{
			desc: "failed to run terraform state list",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "list"},
					exitCode: 1,
				},
			},
			state: nil,
			want:  nil,
			ok:    false,
		},
		{
			desc: "with addresses",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "list", "aws_instance.example", "module.example"},
					stdout:   stdout,
					exitCode: 0,
				},
			},
			state:     nil,
			addresses: []string{"aws_instance.example", "module.example"},
			want:      []string{"aws_security_group.bar", "aws_security_group.foo"},
			ok:        true,
		},
		{
			desc: "with opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "list", "-state=foo.tfstate", "-id=bar"},
					stdout:   stdout,
					exitCode: 0,
				},
			},
			state: nil,
			opts:  []string{"-state=foo.tfstate", "-id=bar"},
			want:  []string{"aws_security_group.bar", "aws_security_group.foo"},
			ok:    true,
		},
		{
			desc: "with addresses and opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "list", "-state=foo.tfstate", "-id=bar", "aws_instance.example", "module.example"},
					stdout:   stdout,
					exitCode: 0,
				},
			},
			state:     nil,
			addresses: []string{"aws_instance.example", "module.example"},
			opts:      []string{"-state=foo.tfstate", "-id=bar"},
			want:      []string{"aws_security_group.bar", "aws_security_group.foo"},
			ok:        true,
		},
		{
			desc: "with state",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "list", "-state=/path/to/tempfile", "-id=bar", "aws_instance.example", "module.example"},
					argsRe:   regexp.MustCompile(`^terraform state list -state=.+ -id=bar aws_instance.example module.example$`),
					stdout:   stdout,
					exitCode: 0,
				},
			},
			state:     &state,
			addresses: []string{"aws_instance.example", "module.example"},
			opts:      []string{"-id=bar"},
			want:      []string{"aws_security_group.bar", "aws_security_group.foo"},
			ok:        true,
		},
		{
			desc: "with state and -state= (conflict error)",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "list", "-state=/path/to/tempfile", "-id=bar", "-state=foo.tfstate", "aws_instance.example", "module.example"},
					argsRe:   regexp.MustCompile(`^terraform state list -state=\S+ -id=bar -state=foo.tfstate aws_instance.example module.example$`),
					exitCode: 1,
				},
			},
			state:     &state,
			addresses: nil,
			opts:      []string{"-id=bar", "-state=foo.tfstate"},
			want:      nil,
			ok:        false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			got, err := terraformCLI.StateList(context.Background(), tc.state, tc.addresses, tc.opts...)
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

const testStateListState = `
{
  "version": 4,
  "terraform_version": "0.12.28",
  "serial": 1,
  "lineage": "a19299f0-68d7-3763-56ca-15ae05f60684",
  "outputs": {},
  "resources": [
    {
      "mode": "managed",
      "type": "aws_security_group",
      "name": "bar",
      "provider": "provider.aws",
      "instances": [
        {
          "schema_version": 1,
          "attributes": {
            "arn": "arn:aws:ec2:ap-northeast-1:000000000000:security-group/sg-ecde6356",
            "description": "Managed by Terraform",
            "egress": [
              {
                "cidr_blocks": [
                  "0.0.0.0/0"
                ],
                "description": "",
                "from_port": 0,
                "ipv6_cidr_blocks": [],
                "prefix_list_ids": [],
                "protocol": "-1",
                "security_groups": [],
                "self": false,
                "to_port": 0
              }
            ],
            "id": "sg-ecde6356",
            "ingress": [],
            "name": "bar",
            "name_prefix": null,
            "owner_id": "000000000000",
            "revoke_rules_on_delete": false,
            "tags": null,
            "timeouts": null,
            "vpc_id": ""
          },
          "private": "eyJlMmJmYjczMC1lY2FhLTExZTYtOGY4OC0zNDM2M2JjN2M0YzAiOnsiY3JlYXRlIjo2MDAwMDAwMDAwMDAsImRlbGV0ZSI6NjAwMDAwMDAwMDAwfSwic2NoZW1hX3ZlcnNpb24iOiIxIn0="
        }
      ]
    },
    {
      "mode": "managed",
      "type": "aws_security_group",
      "name": "foo",
      "provider": "provider.aws",
      "instances": [
        {
          "schema_version": 1,
          "attributes": {
            "arn": "arn:aws:ec2:ap-northeast-1:000000000000:security-group/sg-d1ff4d60",
            "description": "Managed by Terraform",
            "egress": [
              {
                "cidr_blocks": [
                  "0.0.0.0/0"
                ],
                "description": "",
                "from_port": 0,
                "ipv6_cidr_blocks": [],
                "prefix_list_ids": [],
                "protocol": "-1",
                "security_groups": [],
                "self": false,
                "to_port": 0
              }
            ],
            "id": "sg-d1ff4d60",
            "ingress": [],
            "name": "foo",
            "name_prefix": null,
            "owner_id": "000000000000",
            "revoke_rules_on_delete": false,
            "tags": {},
            "timeouts": null,
            "vpc_id": ""
          },
          "private": "eyJlMmJmYjczMC1lY2FhLTExZTYtOGY4OC0zNDM2M2JjN2M0YzAiOnsiY3JlYXRlIjo2MDAwMDAwMDAwMDAsImRlbGV0ZSI6NjAwMDAwMDAwMDAwfSwic2NoZW1hX3ZlcnNpb24iOiIxIn0="
        }
      ]
    }
  ]
}
`
