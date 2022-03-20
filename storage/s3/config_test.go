package s3

import "testing"

func TestConfigNewStorage(t *testing.T) {
	cases := []struct {
		desc   string
		config *Config
		ok     bool
	}{
		{
			desc: "valid",
			config: &Config{
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
				_ = got.(*Storage)
			}
		})
	}
}
