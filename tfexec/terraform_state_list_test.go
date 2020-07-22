package tfexec

import (
	"context"
	"reflect"
	"regexp"
	"testing"
)

func TestTerraformCLIStateList(t *testing.T) {
	state := NewState([]byte(testStateListState))
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
			state:     state,
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
					exitCode: 0,
				},
			},
			state:     state,
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
			if tc.ok && !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got: %v, want: %v", got, tc.want)
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
