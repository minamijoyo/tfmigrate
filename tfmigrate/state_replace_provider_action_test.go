package tfmigrate

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/minamijoyo/tfmigrate/tfexec"
)

func TestAccStateReplaceProviderAction(t *testing.T) {
	tfexec.SkipUnlessAcceptanceTestEnabled(t)

	backend := tfexec.GetTestAccBackendS3Config(t.Name())

	source := `
resource "null_resource" "foo" {}
`

	workspace := "default"
	tf := tfexec.SetupTestAccWithApply(t, workspace, backend+source)
	ctx := context.Background()

	actions := []StateAction{
		NewStateReplaceProviderAction("registry.terraform.io/hashicorp/null", "registry.tfmigrate.io/hashicorp/null"),
	}

	m := NewStateMigrator(tf.Dir(), workspace, actions, &MigratorOption{}, false)
	planErr := m.Plan(ctx)
	if planErr == nil {
		t.Fatalf("expected migrator plan error but received: %s", planErr)
	}

	v, err := tf.Version(ctx)
	if err != nil {
		t.Fatalf("unexpected version error: %s", err)
	}

	constraints, err := version.NewConstraint(fmt.Sprintf(">= %s", tfexec.MinimumTerraformVersionForStateReplaceProvider))
	if err != nil {
		t.Fatalf("unexpected version constraint error: %s", err)
	}

	expected := "Could not load the schema for provider\nregistry.tfmigrate.io/hashicorp/null"
	if !constraints.Check(v) {
		expected = fmt.Sprintf("configuration uses Terraform version %s; replace-provider action requires Terraform version %s", v, constraints)
	}

	if !strings.Contains(planErr.Error(), expected) {
		t.Fatalf("expected migrator plan error to include: %s; received: %s", expected, err)
	}
}
