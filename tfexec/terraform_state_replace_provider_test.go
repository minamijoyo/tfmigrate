package tfexec

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
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

func TestAccTerraformCLIStateReplaceProvider(t *testing.T) {
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

	v, err := terraformCLI.Version(context.Background())
	if err != nil {
		t.Fatalf("unexpected version error: %s", err)
	}

	constraints, err := version.NewConstraint(fmt.Sprintf(">= %s", MinimumTerraformVersionForStateReplaceProvider))
	if err != nil {
		t.Fatalf("unexpected version constraint error: %s", err)
	}

	isLegacyTFVersion := !constraints.Check(v)

	state, err := terraformCLI.StatePull(context.Background())
	if err != nil {
		t.Fatalf("failed to run terraform state pull: %s", err)
	}

	stateProvider := "registry.terraform.io/hashicorp/null"
	if isLegacyTFVersion {
		stateProvider = "provider.null"
	}

	if !strings.Contains(string(state.Bytes()), stateProvider) {
		t.Errorf("state does not contain provider: %s", stateProvider)
	}

	gotProviders, err := terraformCLI.Providers(context.Background())
	if err != nil {
		t.Fatalf("failed to run terraform providers: %s", err)
	}

	wantProviders := providersStdout
	if isLegacyTFVersion {
		wantProviders = legacyProvidersStdout
	}

	if gotProviders != wantProviders {
		t.Errorf("got: %s, want: %s", gotProviders, wantProviders)
	}

	updatedState, err := terraformCLI.StateReplaceProvider(context.Background(), state, "registry.terraform.io/hashicorp/null", "registry.tfmigrate.io/hashicorp/null", "-auto-approve")
	if err != nil && !isLegacyTFVersion {
		t.Fatalf("failed to run terraform state replace-provider: %s", err)
	}

	if err == nil && isLegacyTFVersion {
		t.Fatalf("expected state replace-provider error for legacy Terraform version: %s", v.String())
	}

	if isLegacyTFVersion && !strings.Contains(err.Error(), fmt.Sprintf("configuration uses Terraform version %s; replace-provider action requires Terraform version >= %s", v.String(), MinimumTerraformVersionForStateReplaceProvider)) {
		t.Fatalf("expected state replace-provider error for legacy Terraform version: %s", v.String())
	}

	if !isLegacyTFVersion && !strings.Contains(string(updatedState.Bytes()), "registry.tfmigrate.io/hashicorp/null") {
		t.Errorf("state does not contain updated provider: %s", "registry.tfmigrate.io/hashicorp/null")
	}

	if !isLegacyTFVersion && strings.Contains(string(updatedState.Bytes()), "registry.terraform.io/hashicorp/null") {
		t.Errorf("state contains old provider: %s", "registry.terraform.io/hashicorp/null")
	}
}
