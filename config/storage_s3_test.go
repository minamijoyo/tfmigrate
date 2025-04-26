package config

import (
	"reflect"
	"testing"

	"github.com/minamijoyo/tfmigrate/storage"
	"github.com/minamijoyo/tfmigrate/storage/s3"
)

func TestParseS3StorageBlock(t *testing.T) {
	cases := []struct {
		desc   string
		env    map[string]string
		source string
		want   storage.Config
		ok     bool
	}{
		{
			desc: "valid (required)",
			source: `
tfmigrate {
  history {
    storage "s3" {
      bucket = "tfmigrate-test"
      key    = "tfmigrate/history.json"
    }
  }
}
`,
			want: &s3.Config{
				Bucket: "tfmigrate-test",
				Key:    "tfmigrate/history.json",
			},
			ok: true,
		},
		{
			desc: "valid (with optional)",
			source: `
tfmigrate {
  history {
    storage "s3" {
      bucket = "tfmigrate-test"
      key    = "tfmigrate/history.json"

      region                      = "ap-northeast-1"
      endpoint                    = "http://localstack:4566"
      access_key                  = "dummy"
      secret_key                  = "dummy"
      profile                     = "dev"
      skip_credentials_validation = true
      skip_metadata_api_check     = true
      force_path_style            = true
    }
  }
}
`,
			want: &s3.Config{
				Bucket:                    "tfmigrate-test",
				Key:                       "tfmigrate/history.json",
				Region:                    "ap-northeast-1",
				Endpoint:                  "http://localstack:4566",
				AccessKey:                 "dummy",
				SecretKey:                 "dummy",
				Profile:                   "dev",
				SkipCredentialsValidation: true,
				SkipMetadataAPICheck:      true,
				ForcePathStyle:            true,
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
  history {
    storage "s3" {
      bucket = "tfmigrate-test"
      key    = "tfmigrate/${env.VAR_NAME}/history.json"
    }
  }
}
`,
			want: &s3.Config{
				Bucket: "tfmigrate-test",
				Key:    "tfmigrate/env1/history.json",
			},
			ok: true,
		},
		{
			desc: "missing required attribute (bucket)",
			source: `
tfmigrate {
  history {
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
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			config, err := ParseConfigurationFile("test.hcl", []byte(tc.source))
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
