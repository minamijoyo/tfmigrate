package tfmigrate

import (
	"context"
	"testing"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

func TestAccStateImportAction(t *testing.T) {
	tfexec.SkipUnlessAcceptanceTestEnabled(t)

	backend := tfexec.GetTestAccBackendS3Config(t.Name())

	source := `
resource "time_static" "foo" { triggers = {} }
resource "time_static" "bar" { triggers = {} }
resource "time_static" "baz" { triggers = {} }
`

	workspace := "default"
	tf := tfexec.SetupTestAccWithApply(t, workspace, backend+source)
	ctx := context.Background()

	_, err := tf.StateRm(ctx, nil, []string{"time_static.foo", "time_static.baz"})
	if err != nil {
		t.Fatalf("failed to run terraform state rm: %s", err)
	}

	changed, err := tf.PlanHasChange(ctx, nil)
	if err != nil {
		t.Fatalf("failed to run PlanHasChange: %s", err)
	}
	if !changed {
		t.Fatalf("expect to have changes")
	}

	actions := []StateAction{
		NewStateImportAction("time_static.foo", "2006-01-02T15:04:05Z"),
		NewStateImportAction("time_static.baz", "2006-01-02T15:04:05Z"),
	}

	m := NewStateMigrator(tf.Dir(), workspace, actions, &MigratorOption{}, false)
	err = m.Plan(ctx)
	if err != nil {
		t.Fatalf("failed to run migrator plan: %s", err)
	}

	err = m.Apply(ctx)
	if err != nil {
		t.Fatalf("failed to run migrator apply: %s", err)
	}
}
