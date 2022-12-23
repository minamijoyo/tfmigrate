package tfmigrate

import (
	"context"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-cmp/cmp"
	"github.com/minamijoyo/tfmigrate/tfexec"
)

func TestAccStateMvActionWildcardRenameSecurityGroupResourceNamesFromDocs(t *testing.T) {
	tfexec.SkipUnlessAcceptanceTestEnabled(t)

	backend := tfexec.GetTestAccBackendS3Config(t.Name())

	source := `
resource "null_resource" "foo" {}
resource "null_resource" "bar" {}
`
	tf := tfexec.SetupTestAccWithApply(t, "default", backend+source)
	ctx := context.Background()

	updatedSource := `
resource "null_resource" "foo2" {}
resource "null_resource" "bar2" {}
`
	tfexec.UpdateTestAccSource(t, tf, backend+updatedSource)

	changed, err := tf.PlanHasChange(ctx, nil)
	if err != nil {
		t.Fatalf("failed to run PlanHasChange: %s", err)
	}
	if !changed {
		t.Fatalf("expect to have changes")
	}

	actions := []StateAction{
		NewStateXMvAction("null_resource.*", "null_resource.${1}2"),
	}

	m := NewStateMigrator(tf.Dir(), "default", actions, &MigratorOption{}, false)
	err = m.Plan(ctx)
	if err != nil {
		t.Fatalf("failed to run migrator plan: %s", err)
	}
}

func TestGetNrOfWildcard(t *testing.T) {
	cases := []struct {
		desc        string
		resource    *StateXMvAction
		nrWildcards int
	}{
		{
			desc:        "Simple resource no wildcardChar",
			resource:    NewStateXMvAction("null_resource.foo", "null_resource.foo2"),
			nrWildcards: 0,
		},
		{
			desc:        "Simple wildcardChar for a resource",
			resource:    NewStateXMvAction("null_resource.*", "null_resource.$1"),
			nrWildcards: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.resource.nrOfWildcards()
			if got != tc.nrWildcards {
				t.Errorf("Number of wildcards for %d is not expected %s", got, tc.resource)
			}
		})
	}
}

func TestGetStateMvActionsForStateList(t *testing.T) {
	cases := []struct {
		desc            string
		stateList       []string
		inputXMvAction  *StateXMvAction
		outputMvActions []*StateMvAction
	}{
		{
			desc:      "Simple resource no wildcardChar",
			stateList: nil,
			inputXMvAction: &StateXMvAction{
				source:      "null_resource.foo",
				destination: "null_resource.foo2",
			},
			outputMvActions: []*StateMvAction{
				{
					source:      "null_resource.foo",
					destination: "null_resource.foo2",
				},
			},
		},
		{
			desc:      "Simple resource with wildcardChar",
			stateList: []string{"null_resource.foo"},
			inputXMvAction: &StateXMvAction{
				source:      "null_resource.*",
				destination: "module.example[\"$1\"].this",
			},
			outputMvActions: []*StateMvAction{
				{
					source:      "null_resource.foo",
					destination: "module.example[\"foo\"].this",
				},
			},
		},
		{
			desc:      "Simple module name refactor with wildcardChar",
			stateList: []string{"module.example1[\"foo\"].this"},
			inputXMvAction: &StateXMvAction{
				source:      "module.example1[\"*\"].this",
				destination: "module.example2[\"$1\"].this",
			},
			outputMvActions: []*StateMvAction{
				{
					source:      "module.example1[\"foo\"].this",
					destination: "module.example2[\"foo\"].this",
				},
			},
		},
		{
			desc:      "No matching resources in state",
			stateList: []string{"time_static.foo"},
			inputXMvAction: &StateXMvAction{
				source:      "null_resource.*",
				destination: "module.example[\"$1\"].this",
			},
			outputMvActions: []*StateMvAction{},
		},
		{
			desc:      "Documented feature; positional matching for example to allow switching matches from place",
			stateList: []string{"module[\"bar\"].null_resource.foo"},
			inputXMvAction: &StateXMvAction{
				source:      "module[\"*\"].null_resource.*",
				destination: "module[\"$2\"].null_resource.$1",
			},
			outputMvActions: []*StateMvAction{
				{
					source:      "module[\"bar\"].null_resource.foo",
					destination: "module[\"foo\"].null_resource.bar",
				},
			},
		},
		{
			desc: "Multiple resources refactored into a module",
			stateList: []string{
				"null_resource.foo",
				"null_resource.bar",
				"null_resource.baz",
			},
			inputXMvAction: &StateXMvAction{
				source:      "null_resource.*",
				destination: "module.example[\"$1\"].null_resource.this",
			},
			outputMvActions: []*StateMvAction{
				{
					source:      "null_resource.foo",
					destination: "module.example[\"foo\"].null_resource.this",
				},
				{
					source:      "null_resource.bar",
					destination: "module.example[\"bar\"].null_resource.this",
				},
				{
					source:      "null_resource.baz",
					destination: "module.example[\"baz\"].null_resource.this",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := tc.inputXMvAction.getStateMvActionsForStateList(tc.stateList)
			// Errors are not expected. At this stage the only location from which errors are expected is if the regular
			// expression that comes from the source cannot compile but since meta-characters are quoted and we only
			// introduce matched braces and unmatched meta-characters there are no known cases where we would hit this.
			// Still this case gets handled explicitly as it can be helpful info if the author missed a case.
			if err != nil {
				t.Fatalf("Encountered error %v", err)
			}

			if diff := cmp.Diff(got, tc.outputMvActions, cmp.AllowUnexported(StateMvAction{})); diff != "" {
				t.Errorf("got: %s, want = %s, diff = %s", spew.Sdump(got), spew.Sdump(tc.outputMvActions), diff)
			}

		})
	}
}
