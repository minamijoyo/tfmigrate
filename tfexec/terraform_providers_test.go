package tfexec

import (
	"context"
	"reflect"
	"testing"
)

var providersStdout = `
Providers required by configuration:
.
└── provider[registry.terraform.io/hashicorp/null]

Providers required by state:

    provider[registry.terraform.io/hashicorp/null]

`

var legacyProvidersStdout = `.
└── provider.null

`

func TestTerraformCLIProviders(t *testing.T) {
	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		addresses    []string
		want         string
		ok           bool
	}{
		{
			desc: "basic invocation",
			mockCommands: []*mockCommand{
				{
					stdout:   providersStdout,
					exitCode: 0,
				},
			},
			want: providersStdout,
			ok:   true,
		},
		{
			desc: "failed to run terraform providers",
			mockCommands: []*mockCommand{
				{
					exitCode: 1,
				},
			},
			want: "",
			ok:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			tc.mockCommands[0].args = []string{"terraform", "providers"}
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			got, err := terraformCLI.Providers(context.Background())
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

func TestAccTerraformCLIProviders(t *testing.T) {
	SkipUnlessAcceptanceTestEnabled(t)

	source := `
resource "null_resource" "foo" {}
resource "null_resource" "bar" {}
`
	e := SetupTestAcc(t, source)
	terraformCLI := NewTerraformCLI(e)

	err := terraformCLI.Init(context.Background(), "-input=false", "-no-color")
	if err != nil {
		t.Fatalf("failed to run terraform init: %s", err)
	}

	err = terraformCLI.Apply(context.Background(), nil, "-input=false", "-no-color", "-auto-approve")
	if err != nil {
		t.Fatalf("failed to run terraform apply: %s", err)
	}

	got, err := terraformCLI.Providers(context.Background())
	if err != nil {
		t.Fatalf("failed to run terraform providers: %s", err)
	}

	supportsStateReplaceProvider, _, err := terraformCLI.SupportsStateReplaceProvider(context.Background())
	if err != nil {
		t.Fatalf("failed to determine if Terraform version supports state replace-provider: %s", err)
	}

	want := providersStdout
	if !supportsStateReplaceProvider {
		want = legacyProvidersStdout
	}

	if got != want {
		t.Errorf("got: %s, want: %s", got, want)
	}
}
