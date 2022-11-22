package tfmigrate

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

func TestAccStateMvActionWildcardRenameSecurityGroupResourceNamesFromDocs(t *testing.T) {
	tfexec.SkipUnlessAcceptanceTestEnabled(t)

	backend := tfexec.GetTestAccBackendS3Config(t.Name())

	source := `
resource "aws_security_group" "foo" {}
resource "aws_security_group" "bar" {}
`
	tf := tfexec.SetupTestAccWithApply(t, "default", backend+source)
	ctx := context.Background()

	updatedSource := `
resource "aws_security_group" "foo2" {}
resource "aws_security_group" "bar2" {}
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
		NewStateXMvAction("aws_security_group.*", "aws_security_group.${1}2"),
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
		{
			desc:      "Documented feature; positional matching for example to allow switching matches from place",
			stateList: []string{"module[\"bar\"].aws_vpc.foo"},
			inputXMvAction: &StateXMvAction{
				source:      "module[\"*\"].aws_vpc.*",
				destination: "module[\"$2\"].aws_vpc.$1",
			},
			outputMvActions: []*StateMvAction{
				{
					source:      "module[\"bar\"].aws_vpc.foo",
					destination: "module[\"foo\"].aws_vpc.bar",
				},
			},
		},
		{
			desc: "Multiple resources refactored into a module",
			stateList: []string{
				"aws_security_group.foo",
				"aws_security_group.bar",
				"aws_security_group.baz",
			},
			inputXMvAction: &StateXMvAction{
				source:      "aws_security_group.*",
				destination: "module.sg_module[\"$1\"].aws_security_group.sg",
			},
			outputMvActions: []*StateMvAction{
				{
					source:      "aws_security_group.foo",
					destination: "module.sg_module[\"foo\"].aws_security_group.sg",
				},
				{
					source:      "aws_security_group.bar",
					destination: "module.sg_module[\"bar\"].aws_security_group.sg",
				},
				{
					source:      "aws_security_group.baz",
					destination: "module.sg_module[\"baz\"].aws_security_group.sg",
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
				t.Errorf("Encountered error %v", err)
			}

			if !reflect.DeepEqual(got, tc.outputMvActions) {
				t.Errorf("got: %v, expected: %v", got, tc.outputMvActions)
			}

		})
	}
}

// At this time purely for testing purposes if a real string function is needed for code this can be replaced.
func (a StateMvAction) String() string {
	return fmt.Sprintf("mv %s %s", a.source, a.destination)
}
