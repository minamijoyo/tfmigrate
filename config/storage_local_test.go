package config

import (
	"reflect"
	"testing"

	"github.com/minamijoyo/tfmigrate/history"
)

func TestLocalStorageConfigNewStorage(t *testing.T) {
	cases := []struct {
		desc   string
		config *LocalStorageConfig
		ok     bool
	}{
		{
			desc: "valid",
			config: &LocalStorageConfig{
				Path: "tmp/history.json",
			},
			ok: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := tc.config.NewStorage()
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error, got: %#v", got)
			}
			if tc.ok {
				_ = got.(*history.LocalStorage)
			}
		})
	}
}

func TestParseLocalStorageBlock(t *testing.T) {
	cases := []struct {
		desc   string
		source string
		want   StorageConfig
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
			want: &LocalStorageConfig{
				Path: "tmp/history.json",
			},
			ok: true,
		},
		{
			desc: "missing required attribute (path)",
			source: `
tfmigrate {
  history {
    migration_dir = "tfmigrate"
    storage "local" {
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
