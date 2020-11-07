package config

import (
	"reflect"
	"testing"

	"github.com/minamijoyo/tfmigrate/history"
)

func TestParseHistoryBlock(t *testing.T) {
	cases := []struct {
		desc   string
		source string
		want   *history.Config
		ok     bool
	}{
		{
			desc: "valid",
			source: `
tfmigrate {
  history {
    migration_dir = "tfmigrate"
    storage "local" {
      path = "tmp/history.json"
    }
  }
}
`,
			want: &history.Config{
				MigrationDir: "tfmigrate",
				Storage: &history.LocalStorageConfig{
					Path: "tmp/history.json",
				},
			},
			ok: true,
		},
		{
			desc: "missing attribute (migration_dir)",
			source: `
tfmigrate {
  history {
    storage "local" {
      path = "tmp/history.json"
    }
  }
}
`,
			want: nil,
			ok:   false,
		},
		{
			desc: "missing block (storage)",
			source: `
tfmigrate {
  history {
    migration_dir = "tfmigrate"
  }
}
`,
			want: nil,
			ok:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			config, err := ParseConfigurationFile("test.hcl", []byte(tc.source))
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error, got: %#v", config)
			}
			if tc.ok {
				got := config.History
				if !reflect.DeepEqual(got, tc.want) {
					t.Errorf("got: %#v, want: %#v", got, tc.want)
				}
			}
		})
	}
}
