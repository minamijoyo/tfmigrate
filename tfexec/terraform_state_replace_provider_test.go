package tfexec

import (
	"context"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestTerraformCLIStateReplaceProvider(t *testing.T) {
	state := NewState([]byte("dummy state"))
	updatedState := NewState([]byte("updated dummy state"))

	// mock writing state to a temporary file.
	runFunc := func(args ...string) error {
		for _, arg := range args {
			if strings.HasPrefix(arg, "-state=") {
				stateFile := arg[len("-state="):]
				err := os.WriteFile(stateFile, updatedState.Bytes(), 0600)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	versionRunFunc := func(args ...string) error {
		return nil
	}

	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		state        *State
		source       string
		destination  string
		opts         []string
		updatedState *State
		ok           bool
	}{
		{
			desc: "no opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "version"},
					argsRe:   regexp.MustCompile(`^terraform version$`),
					runFunc:  versionRunFunc,
					stdout:   "Terraform v1.5.0\n",
					exitCode: 0,
				},
				{
					args:     []string{"terraform", "state", "replace-provider", "registry.terraform.io/-/null", "registry.terraform.io/hashicorp/null"},
					argsRe:   regexp.MustCompile(`^terraform state replace-provider registry.terraform.io/-/null registry.terraform.io/hashicorp/null$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:        nil,
			source:       "registry.terraform.io/-/null",
			destination:  "registry.terraform.io/hashicorp/null",
			updatedState: nil,
			ok:           true,
		},
		{
			desc: "failed to run terraform state replace-provider",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "version"},
					argsRe:   regexp.MustCompile(`^terraform version$`),
					runFunc:  versionRunFunc,
					stdout:   "Terraform v1.5.0\n",
					exitCode: 0,
				},
				{
					args:     []string{"terraform", "state", "replace-provider", "registry.terraform.io/-/null", "registry.terraform.io/hashicorp/null"},
					argsRe:   regexp.MustCompile(`^terraform state replace-provider registry.terraform.io/-/null registry.terraform.io/hashicorp/null$`),
					runFunc:  runFunc,
					exitCode: 1,
				},
			},
			state:        nil,
			source:       "registry.terraform.io/-/null",
			destination:  "registry.terraform.io/hashicorp/null",
			updatedState: nil,
			ok:           false,
		},
		{
			desc: "with unsupported Terraform version",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "version"},
					argsRe:   regexp.MustCompile(`^terraform version$`),
					runFunc:  versionRunFunc,
					stdout:   "Terraform v0.12.99\n",
					exitCode: 0,
				},
			},
			state:        nil,
			source:       "registry.terraform.io/-/null",
			destination:  "registry.terraform.io/hashicorp/null",
			updatedState: nil,
			ok:           false,
		},
		{
			desc: "when version check fails",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "version"},
					argsRe:   regexp.MustCompile(`^terraform version$`),
					runFunc:  versionRunFunc,
					stdout:   "Terraform v1.4.7\n",
					exitCode: 1,
				},
			},
			state:        nil,
			source:       "registry.terraform.io/-/null",
			destination:  "registry.terraform.io/hashicorp/null",
			updatedState: nil,
			ok:           false,
		},
		{
			desc: "with opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "version"},
					argsRe:   regexp.MustCompile(`^terraform version$`),
					runFunc:  versionRunFunc,
					stdout:   "Terraform v1.5.0\n",
					exitCode: 0,
				},
				{
					args:     []string{"terraform", "state", "replace-provider", "-lock=true", "-lock-timeout=10s", "registry.terraform.io/-/null", "registry.terraform.io/hashicorp/null"},
					argsRe:   regexp.MustCompile(`^terraform state replace-provider -lock=true -lock-timeout=10s registry.terraform.io/-/null registry.terraform.io/hashicorp/null$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:        nil,
			source:       "registry.terraform.io/-/null",
			destination:  "registry.terraform.io/hashicorp/null",
			opts:         []string{"-lock=true", "-lock-timeout=10s"},
			updatedState: nil,
			ok:           true,
		},
		{
			desc: "with state",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "version"},
					argsRe:   regexp.MustCompile(`^terraform version$`),
					runFunc:  versionRunFunc,
					stdout:   "Terraform v1.5.0\n",
					exitCode: 0,
				},
				{
					args:     []string{"terraform", "state", "replace-provider", "-state=/path/to/tempfile", "-lock=true", "-lock-timeout=10s", "registry.terraform.io/-/null", "registry.terraform.io/hashicorp/null"},
					argsRe:   regexp.MustCompile(`^terraform state replace-provider -state=.+ -lock=true -lock-timeout=10s registry.terraform.io/-/null registry.terraform.io/hashicorp/null$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:        state,
			source:       "registry.terraform.io/-/null",
			destination:  "registry.terraform.io/hashicorp/null",
			opts:         []string{"-lock=true", "-lock-timeout=10s"},
			updatedState: updatedState,
			ok:           true,
		},
		{
			desc: "with state and -state= (conflict error)",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "version"},
					argsRe:   regexp.MustCompile(`^terraform version$`),
					runFunc:  versionRunFunc,
					stdout:   "Terraform v1.5.0\n",
					exitCode: 0,
				},
				{
					args:     []string{"terraform", "state", "replace-provider", "-state=/path/to/tempfile", "-lock=true", "-state=foo.tfstate", "registry.terraform.io/-/null", "registry.terraform.io/hashicorp/null"},
					argsRe:   regexp.MustCompile(`^terraform state replace-provider -state=.+ -lock=true -state=foo.tfstate registry.terraform.io/-/null registry.terraform.io/hashicorp/null$`),
					runFunc:  runFunc,
					exitCode: 0,
				},
			},
			state:        state,
			source:       "registry.terraform.io/-/null",
			destination:  "registry.terraform.io/hashicorp/null",
			opts:         []string{"-lock=true", "-state=foo.tfstate"},
			updatedState: nil,
			ok:           false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			terraformCLI.SetExecPath("terraform")
			gotState, err := terraformCLI.StateReplaceProvider(context.Background(), tc.state, tc.source, tc.destination, tc.opts...)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}
			if tc.ok {
				if tc.updatedState != nil {
					if !reflect.DeepEqual(gotState.Bytes(), tc.updatedState.Bytes()) {
						t.Errorf("got state: %v, want state: %v", gotState, tc.updatedState)
					}
				} else { // tc.updateState == nil
					if gotState != nil {
						t.Errorf("got state: %v, want state: %v", gotState, tc.updatedState)
					}
				}
			}
		})
	}
}

func TestAccTerraformCLIStateReplaceProviderWithLegacyTerraform(t *testing.T) {
	tfVersion := os.Getenv("TERRAFORM_VERSION")
	if tfVersion != LegacyTerraformVersion {
		t.Skipf("skip %s acceptance test for non-legacy Terraform version %s", t.Name(), tfVersion)
	}

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

	state, err := terraformCLI.StatePull(context.Background())
	if err != nil {
		t.Fatalf("failed to run terraform state pull: %s", err)
	}

	stateProvider := "provider.null"

	if !strings.Contains(string(state.Bytes()), stateProvider) {
		t.Errorf("state does not contain provider: %s", stateProvider)
	}

	gotProviders, err := terraformCLI.Providers(context.Background())
	if err != nil {
		t.Fatalf("failed to run terraform providers: %s", err)
	}

	wantProviders := legacyTerraformProvidersStdout

	if gotProviders != wantProviders {
		t.Errorf("got: %s, want: %s", gotProviders, wantProviders)
	}

	_, err = terraformCLI.StateReplaceProvider(context.Background(), state, "registry.terraform.io/hashicorp/null", "registry.tfmigrate.io/hashicorp/null", "-auto-approve")
	if err == nil {
		t.Fatalf("expected terraform state replace-provider to error")
	}

	expected := "replace-provider action requires Terraform version >= 0.13"
	if err.Error() != expected {
		t.Fatalf("expected terraform state replace-provider to error with %s; got %s", expected, err.Error())
	}
}

func TestAccTerraformCLIStateReplaceProvider(t *testing.T) {
	SkipUnlessAcceptanceTestEnabled(t)

	tfVersion := os.Getenv("TERRAFORM_VERSION")
	if tfVersion == LegacyTerraformVersion {
		t.Skipf("skip %s acceptance test for legacy Terraform version %s", t.Name(), tfVersion)
	}

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

	state, err := terraformCLI.StatePull(context.Background())
	if err != nil {
		t.Fatalf("failed to run terraform state pull: %s", err)
	}

	execType, _, err := terraformCLI.Version(context.Background())
	if err != nil {
		t.Fatalf("failed to detect execType: %s", err)
	}
	fromProviderFQN := ""
	switch execType {
	case "terraform":
		fromProviderFQN = "registry.terraform.io/hashicorp/null"
	case "opentofu":
		fromProviderFQN = "registry.opentofu.org/hashicorp/null"
	}

	if !strings.Contains(string(state.Bytes()), fromProviderFQN) {
		t.Errorf("state does not contain provider: %s", fromProviderFQN)
	}

	gotProviders, err := terraformCLI.Providers(context.Background())
	if err != nil {
		t.Fatalf("failed to run terraform providers: %s", err)
	}

	wantProviders := ""
	switch execType {
	case "terraform":
		wantProviders = terraformProvidersStdout
	case "opentofu":
		wantProviders = opentofuProvidersStdout
	}

	if gotProviders != wantProviders {
		t.Errorf("got: %s, want: %s", gotProviders, wantProviders)
	}

	updatedState, err := terraformCLI.StateReplaceProvider(context.Background(), state, fromProviderFQN, "registry.tfmigrate.io/hashicorp/null", "-auto-approve")
	if err != nil {
		t.Fatalf("failed to run terraform state replace-provider: %s", err)
	}

	if !strings.Contains(string(updatedState.Bytes()), "registry.tfmigrate.io/hashicorp/null") {
		t.Errorf("state does not contain updated provider: %s", "registry.tfmigrate.io/hashicorp/null")
	}

	if strings.Contains(string(updatedState.Bytes()), fromProviderFQN) {
		t.Errorf("state contains old provider: %s", fromProviderFQN)
	}
}
