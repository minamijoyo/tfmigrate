package tfmigrate

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-cmp/cmp"
)

func TestGetNrOfWildcard(t *testing.T) {
	cases := []struct {
		desc   string
		action *StateXmvAction
		want   int
	}{
		{
			desc:   "simple resource no wildcardChar",
			action: NewStateXmvAction("null_resource.foo", "null_resource.foo2"),
			want:   0,
		},
		{
			desc:   "simple wildcardChar for a resource",
			action: NewStateXmvAction("null_resource.*", "null_resource.$1"),
			want:   1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := newXmvExpander(tc.action)
			got := e.nrOfWildcards()
			if got != tc.want {
				t.Errorf("got: %d, but want: %d", got, tc.want)
			}
		})
	}
}

func TestXmvExpanderExpand(t *testing.T) {
	cases := []struct {
		desc            string
		stateList       []string
		inputXMvAction  *StateXmvAction
		outputMvActions []*StateMvAction
	}{
		{
			desc:      "simple resource no wildcardChar",
			stateList: nil,
			inputXMvAction: &StateXmvAction{
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
			desc:      "simple resource with wildcardChar",
			stateList: []string{"null_resource.foo"},
			inputXMvAction: &StateXmvAction{
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
			desc:      "simple module name refactor with wildcardChar",
			stateList: []string{"module.example1[\"foo\"].this"},
			inputXMvAction: &StateXmvAction{
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
			desc:      "no matching resources in state",
			stateList: []string{"time_static.foo"},
			inputXMvAction: &StateXmvAction{
				source:      "null_resource.*",
				destination: "module.example[\"$1\"].this",
			},
			outputMvActions: []*StateMvAction{},
		},
		{
			desc:      "documented feature; positional matching for example to allow switching matches from place",
			stateList: []string{"module[\"bar\"].null_resource.foo"},
			inputXMvAction: &StateXmvAction{
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
			desc: "multiple resources refactored into a module",
			stateList: []string{
				"null_resource.foo",
				"null_resource.bar",
				"null_resource.baz",
			},
			inputXMvAction: &StateXmvAction{
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
		{
			desc: "match all for merging multiple tfstates",
			stateList: []string{
				"null_resource.foo",
				"null_resource.bar",
			},
			inputXMvAction: &StateXmvAction{
				source:      "*",
				destination: "$1",
			},
			outputMvActions: []*StateMvAction{
				{
					source:      "null_resource.foo",
					destination: "null_resource.foo",
				},
				{
					source:      "null_resource.bar",
					destination: "null_resource.bar",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := newXmvExpander(tc.inputXMvAction)
			got, err := e.expand(tc.stateList)
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
