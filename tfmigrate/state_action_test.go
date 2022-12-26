package tfmigrate

import (
	"reflect"
	"testing"
)

func TestNewStateActionFromString(t *testing.T) {
	cases := []struct {
		desc   string
		cmdStr string
		want   StateAction
		ok     bool
	}{
		{
			desc:   "mv action (valid)",
			cmdStr: "mv null_resource.foo null_resource.foo2",
			want: &StateMvAction{
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
			want: &StateXmvAction{
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
			desc:   "rm action (valid)",
			cmdStr: "rm time_static.foo",
			want: &StateRmAction{
				addresses: []string{"time_static.foo"},
			},
			ok: true,
		},
		{
			desc:   "rm action (no args)",
			cmdStr: "rm",
			want:   nil,
			ok:     false,
		},
		{
			desc:   "rm action (2 args)",
			cmdStr: "rm time_static.foo time_static.bar",
			want: &StateRmAction{
				addresses: []string{"time_static.foo", "time_static.bar"},
			},
			ok: true,
		},
		{
			desc:   "import action (valid)",
			cmdStr: "import time_static.foo 2006-01-02T15:04:05Z",
			want: &StateImportAction{
				address: "time_static.foo",
				id:      "2006-01-02T15:04:05Z",
			},
			ok: true,
		},
		{
			desc:   "import action (no args)",
			cmdStr: "import",
			want:   nil,
			ok:     false,
		},
		{
			desc:   "import action (1 arg)",
			cmdStr: "import time_static.foo",
			want:   nil,
			ok:     false,
		},
		{
			desc:   "import action (3 args)",
			cmdStr: "import time_static.foo foo bar",
			want:   nil,
			ok:     false,
		},
		{
			desc:   "duplicated white spaces",
			cmdStr: " mv  null_resource.foo    null_resource.foo2 ",
			want: &StateMvAction{
				source:      "null_resource.foo",
				destination: "null_resource.foo2",
			},
			ok: true,
		},
		{
			desc:   "a white space in resource address",
			cmdStr: `mv docker_container.nginx 'docker_container.nginx["This is an example"]'`,
			want: &StateMvAction{
				source:      "docker_container.nginx",
				destination: `docker_container.nginx["This is an example"]`,
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
			got, err := NewStateActionFromString(tc.cmdStr)
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

func TestSplitStateAction(t *testing.T) {
	cases := []struct {
		desc   string
		cmdStr string
		want   []string
		ok     bool
	}{
		{
			desc:   "simple",
			cmdStr: "mv null_resource.foo null_resource.foo2",
			want:   []string{"mv", "null_resource.foo", "null_resource.foo2"},
			ok:     true,
		},
		{
			desc:   "duplicated white spaces",
			cmdStr: " mv  null_resource.foo    null_resource.foo2 ",
			want:   []string{"mv", "null_resource.foo", "null_resource.foo2"},
			ok:     true,
		},
		{
			desc:   "a white space in resource address",
			cmdStr: `mv docker_container.nginx 'docker_container.nginx["This is an example"]'`,
			want:   []string{"mv", "docker_container.nginx", `docker_container.nginx["This is an example"]`},
			ok:     true,
		},
		{
			desc:   "syntax error (unmatch quote)",
			cmdStr: `mv foo 'bar`,
			want:   nil,
			ok:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := splitStateAction(tc.cmdStr)
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
