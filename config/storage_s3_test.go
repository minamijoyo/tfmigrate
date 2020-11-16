package config

import (
	"reflect"
	"testing"

	"github.com/minamijoyo/tfmigrate/history"
)

func TestParseS3StorageBlock(t *testing.T) {
	cases := []struct {
		desc   string
		source string
		want   history.StorageConfig
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
			want: &history.S3StorageConfig{
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
			want: &history.S3StorageConfig{
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
