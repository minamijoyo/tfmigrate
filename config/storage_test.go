package config

import (
	"reflect"
	"testing"

	"github.com/minamijoyo/tfmigrate/history"
)

func TestParseStorageBlock(t *testing.T) {
	cases := []struct {
		desc   string
		source string
		want   history.StorageConfig
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
			want: &history.LocalStorageConfig{
				Path: "tmp/history.json",
			},
			ok: true,
		},
		{
			desc: "unknown type",
			source: `
tfmigrate {
  history {
    migration_dir = "tfmigrate"
    storage "foo" {
    }
  }
}
`,
			want: nil,
			ok:   false,
		},
		{
			desc: "missing type",
			source: `
tfmigrate {
  history {
    migration_dir = "tfmigrate"
    storage {
    }
  }
}
`,
			want: nil,
			ok:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			config, err := ParseSettingFile("test.hcl", []byte(tc.source))
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error, got: %#v", config)
			}
			if tc.ok {
				got := config.History.Storage
				if !reflect.DeepEqual(got, tc.want) {
					t.Errorf("got: %#v, want: %#v", got, tc.want)
				}
			}
		})
	}
}
