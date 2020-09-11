package config

import (
	"reflect"
	"testing"
)

func TestParseSettingFile(t *testing.T) {
	cases := []struct {
		desc   string
		source string
		want   *TfmigrateConfig
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
			want: &TfmigrateConfig{
				History: &HistoryConfig{
					MigrationDir: "tfmigrate",
					Storage: &LocalStorageConfig{
						Path: "tmp/history.json",
					},
				},
			},
			ok: true,
		},
		{
			desc: "missing block (history)",
			source: `
tfmigrate {
}
`,
			want: &TfmigrateConfig{},
			ok:   true,
		},
		{
			desc: "unknown block",
			source: `
foo {
}
`,
			want: nil,
			ok:   false,
		},
		{
			desc:   "empty file",
			source: ``,
			want:   nil,
			ok:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := ParseSettingFile("test.hcl", []byte(tc.source))
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
