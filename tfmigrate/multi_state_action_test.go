package tfmigrate

import (
	"reflect"
	"testing"
)

func TestNewMultiStateActionFromString(t *testing.T) {
	cases := []struct {
		desc   string
		cmdStr string
		want   MultiStateAction
		ok     bool
	}{
		{
			desc:   "mv action (valid)",
			cmdStr: "mv null_resource.foo null_resource.foo2",
			want: &MultiStateMvAction{
				source:      "null_resource.foo",
				destination: "null_resource.foo2",
			},
			ok: true,
		},
		{
			desc:   "mv action (no args)",
			cmdStr: "mv",
			want:   nil,
			ok:     false,
		},
		{
			desc:   "mv action (1 arg)",
			cmdStr: "mv null_resource.foo",
			want:   nil,
			ok:     false,
		},
		{
			desc:   "mv action (3 args)",
			cmdStr: "mv null_resource.foo null_resource.foo2  null_resource.foo3",
			want:   nil,
			ok:     false,
		},
		{
			desc:   "xmv action (valid)",
			cmdStr: "xmv null_resource.* null_resource.$1",
			want: &MultiStateXmvAction{
				source:      "null_resource.*",
				destination: "null_resource.$1",
			},
			ok: true,
		},
		{
			desc:   "xmv action (no args)",
			cmdStr: "xmv",
			want:   nil,
			ok:     false,
		},
		{
			desc:   "xmv action (1 arg)",
			cmdStr: "xmv null_resource.foo",
			want:   nil,
			ok:     false,
		},
		{
			desc:   "xmv action (3 args)",
			cmdStr: "xmv null_resource.foo null_resource.foo2  null_resource.foo3",
			want:   nil,
			ok:     false,
		},
		{
			desc:   "duplicated white spaces",
			cmdStr: " mv  null_resource.foo    null_resource.foo2 ",
			want: &MultiStateMvAction{
				source:      "null_resource.foo",
				destination: "null_resource.foo2",
			},
			ok: true,
		},
		{
			desc:   "unknown type",
			cmdStr: "foo bar baz",
			want:   nil,
			ok:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := NewMultiStateActionFromString(tc.cmdStr)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error, got: %#v", got)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got: %#v, want: %#v", got, tc.want)
			}
		})
	}
}
