package config

import (
	"reflect"
	"testing"

	"github.com/minamijoyo/tfmigrate/history"
	"github.com/minamijoyo/tfmigrate/storage/local"
)

func TestParseConfigurationFile(t *testing.T) {
	cases := []struct {
		desc   string
		env    map[string]string
		source string
		want   *TfmigrateConfig
		ok     bool
	}{
		{
			desc: "valid",
			env:  nil,
			source: `
tfmigrate {
  migration_dir = "tfmigrate"
  history {
    storage "local" {
      path = "tmp/history.json"
    }
  }
}
`,
			want: &TfmigrateConfig{
				ExecPath:     "terraform",
				MigrationDir: "tfmigrate",
				History: &history.Config{
					Storage: &local.Config{
						Path: "tmp/history.json",
					},
				},
			},
			ok: true,
		},
		{
			desc: "default migration_dir",
			env:  nil,
			source: `
tfmigrate {
  history {
    storage "local" {
      path = "tmp/history.json"
    }
  }
}
`,
			want: &TfmigrateConfig{
				ExecPath:     "terraform",
				MigrationDir: ".",
				History: &history.Config{
					Storage: &local.Config{
						Path: "tmp/history.json",
					},
				},
			},
			ok: true,
		},
		{
			desc: "env vars",
			env: map[string]string{
				"VAR_NAME": "env1",
			},
			source: `
tfmigrate {
  migration_dir = "tfmigrate/${env.VAR_NAME}"
  history {
    storage "local" {
      path = "tmp/${env.VAR_NAME}/history.json"
    }
  }
}
`,
			want: &TfmigrateConfig{
				ExecPath:     "terraform",
				MigrationDir: "tfmigrate/env1",
				History: &history.Config{
					Storage: &local.Config{
						Path: "tmp/env1/history.json",
					},
				},
			},
			ok: true,
		},
		{
			desc: "missing block (history)",
			env:  nil,
			source: `
tfmigrate {
}
`,
			want: &TfmigrateConfig{
				ExecPath:     "terraform",
				MigrationDir: ".",
				History:      nil,
			},
			ok: true,
		},
		{
			desc: "unknown block",
			env:  nil,
			source: `
foo {
}
`,
			want: nil,
			ok:   false,
		},
		{
			desc:   "empty file",
			env:    nil,
			source: ``,
			want:   nil,
			ok:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			got, err := ParseConfigurationFile("test.hcl", []byte(tc.source))
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error, got: %#v", got)
			}
			if tc.ok {
				if !reflect.DeepEqual(got, tc.want) {
					t.Errorf("got: %#v, want: %#v", got, tc.want)
				}
			}
		})
	}
}
