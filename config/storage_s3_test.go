package config

import (
	"reflect"
	"testing"

	"github.com/minamijoyo/tfmigrate/history"
)

func TestS3StorageConfigNewStorage(t *testing.T) {
	cases := []struct {
		desc   string
		config *S3StorageConfig
		ok     bool
	}{
		{
			desc: "valid",
			config: &S3StorageConfig{
				Bucket: "tfmigrate-test",
				Key:    "tfmigrate/history.json",
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
				_ = got.(*history.S3Storage)
			}
		})
	}
}

func TestParseS3StorageBlock(t *testing.T) {
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
    storage "s3" {
      bucket = "tfmigrate-test"
      key    = "tfmigrate/history.json"
    }
  }
}
`,
			want: &S3StorageConfig{
				Bucket: "tfmigrate-test",
				Key:    "tfmigrate/history.json",
			},
			ok: true,
		},
		{
			desc: "missing required attribute (bucket)",
			source: `
tfmigrate {
  history {
    migration_dir = "tfmigrate"
    storage "s3" {
      key    = "tfmigrate/history.json"
    }
  }
}
`,
			want: nil,
			ok:   false,
		},
		{
			desc: "missing required attribute (key)",
			source: `
tfmigrate {
  history {
    migration_dir = "tfmigrate"
    storage "s3" {
      bucket = "tfmigrate-test"
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
