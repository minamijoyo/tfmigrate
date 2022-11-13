package tfmigrate

import (
	"context"
	"testing"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

func TestAccStateMvActionWildcardRefactorToModule(t *testing.T) {
	tfexec.SkipUnlessAcceptanceTestEnabled(t)

	backend := tfexec.GetTestAccBackendS3Config(t.Name())

	source := `
resource "aws_security_group" "foo" {}
resource "aws_security_group" "bar" {}
resource "aws_security_group" "baz" {}
`
	tf := tfexec.SetupTestAccWithApply(t, "default", backend+source)
	ctx := context.Background()

	updatedSource := `
module "sg_module" {
 source = "./sg_module"
 for_each = toset(["foo", "bar", "baz"])
}
`
	sgModuleSource := `
resource "aws_security_group" "sg" {}
`

	tfexec.AddModuleToTestAcc(t, tf, "sg_module", sgModuleSource)
	tfexec.UpdateTestAccSource(t, tf, backend+updatedSource)
	tfexec.RunTestAccInitModule(t, tf)

	changed, err := tf.PlanHasChange(ctx, nil)
	if err != nil {
		t.Fatalf("failed to run PlanHasChange: %s", err)
	}
	if !changed {
		t.Fatalf("expect to have changes")
	}

	actions := []StateAction{
		NewStateXMvAction("aws_security_group.*", "module.sg_module[\"$1\"].aws_security_group.sg"),
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
			resource:    NewStateXMvAction("aws_security_group.foo", "aws_security_group.foo2"),
			nrWildcards: 0,
		},
		{
			desc:        "Simple wildcardChar for a resource",
			resource:    NewStateXMvAction("aws_security_group.*", "aws_security_group.$1"),
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

func TestSingleResourceStateMoves(t *testing.T) {
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
				source:      "aws_security_group.foo",
				destination: "aws_security_group.foo2",
			},
			outputMvActions: []*StateMvAction{
				{
					source:      "aws_security_group.foo",
					destination: "aws_security_group.foo2",
				},
			},
		},
		{
			desc:      "Simple resource with wildcardChar",
			stateList: []string{"aws_security_group.foo"},
			inputXMvAction: &StateXMvAction{
				source:      "aws_security_group.*",
				destination: "module.security_group[\"$1\"].sg",
			},
			outputMvActions: []*StateMvAction{
				{
					source:      "aws_security_group.foo",
					destination: "module.security_group[\"foo\"].sg",
				},
			},
		},
		{
			desc:      "Simple module name refactor with wildcardChar",
			stateList: []string{"module.security_group[\"foo\"].sg"},
			inputXMvAction: &StateXMvAction{
				source:      "module.security_group[\"*\"].sg",
				destination: "module.sg[\"$1\"].sg",
			},
			outputMvActions: []*StateMvAction{
				{
					source:      "module.security_group[\"foo\"].sg",
					destination: "module.sg[\"foo\"].sg",
				},
			},
		},
		{
			desc:      "No matching resources in state",
			stateList: []string{"aws_vpc.foo"},
			inputXMvAction: &StateXMvAction{
				source:      "aws_security_group.*",
				destination: "module.security_group[\"$1\"].sg",
			},
			outputMvActions: []*StateMvAction{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := tc.inputXMvAction.getStateMvActionsForStateList(tc.stateList)
			// Should be location agnostic
			if err != nil {
				t.Errorf("Encountered error %v", err)
			}
			if len(got) != len(tc.outputMvActions) {
				t.Errorf("Did not get same amount of matches: %d, %d", len(got), len(tc.outputMvActions))
			}

			for i, outputMvAction := range got {
				if tc.outputMvActions[i].source != outputMvAction.source {
					t.Errorf("Mistmatching output sources %d: %v, %v", i, tc.outputMvActions[i].source, outputMvAction.source)
				}
				if tc.outputMvActions[i].destination != outputMvAction.destination {
					t.Errorf("Mistmatching output destination %d: %v, %v", i, tc.outputMvActions[i].destination, outputMvAction.destination)
				}
			}
		})
	}
}
