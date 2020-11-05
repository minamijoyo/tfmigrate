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
			cmdStr: "mv aws_security_group.foo aws_security_group.foo2",
			want: &StateMvAction{
				source:      "aws_security_group.foo",
				destination: "aws_security_group.foo2",
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
			cmdStr: "mv aws_security_group.foo",
			want:   nil,
			ok:     false,
		},
		{
			desc:   "mv action (3 args)",
			cmdStr: "mv aws_security_group.foo aws_security_group.foo2  ws_security_group.foo3",
			want:   nil,
			ok:     false,
		},
		{
			desc:   "rm action (valid)",
			cmdStr: "rm aws_security_group.foo",
			want: &StateRmAction{
				addresses: []string{"aws_security_group.foo"},
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
			cmdStr: "rm aws_security_group.foo aws_security_group.bar",
			want: &StateRmAction{
				addresses: []string{"aws_security_group.foo", "aws_security_group.bar"},
			},
			ok: true,
		},
		{
			desc:   "import action (valid)",
			cmdStr: "import aws_security_group.foo foo",
			want: &StateImportAction{
				address: "aws_security_group.foo",
				id:      "foo",
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
			cmdStr: "import aws_security_group.foo",
			want:   nil,
			ok:     false,
		},
		{
			desc:   "import action (3 args)",
			cmdStr: "import aws_security_group.foo foo bar",
			want:   nil,
			ok:     false,
		},
		{
			desc:   "duplicated white spaces",
			cmdStr: " mv  aws_security_group.foo    aws_security_group.foo2 ",
			want: &StateMvAction{
				source:      "aws_security_group.foo",
				destination: "aws_security_group.foo2",
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
			cmdStr: "mv aws_security_group.foo aws_security_group.foo2",
			want:   []string{"mv", "aws_security_group.foo", "aws_security_group.foo2"},
			ok:     true,
		},
		{
			desc:   "duplicated white spaces",
			cmdStr: " mv  aws_security_group.foo    aws_security_group.foo2 ",
			want:   []string{"mv", "aws_security_group.foo", "aws_security_group.foo2"},
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
