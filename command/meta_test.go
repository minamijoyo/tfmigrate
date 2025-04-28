package command

import (
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origWD); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}()

	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	defaultPath := ".tfmigrate.hcl"

	cases := []struct {
		desc      string
		filename  string
		setupEnv  func()
		setupFile func()
		wantDir   string
		wantErr   bool
	}{
		{
			desc:     "config arg, env unset",
			filename: "arg.hcl",
			setupEnv: func() {},
			setupFile: func() {
				content := `
tfmigrate {
  migration_dir = "arg"
}
`
				if err := os.WriteFile("arg.hcl", []byte(content), 0600); err != nil {
					t.Fatal(err)
				}
			},
			wantDir: "arg",
			wantErr: false,
		},
		{
			desc:     "no arg, env set",
			filename: defaultPath,
			setupEnv: func() { os.Setenv("TFMIGRATE_CONFIG", "env.hcl") },
			setupFile: func() {
				content := `
tfmigrate {
  migration_dir = "env"
}
`
				if err := os.WriteFile("env.hcl", []byte(content), 0600); err != nil {
					t.Fatal(err)
				}
			},
			wantDir: "env",
			wantErr: false,
		},
		{
			desc:     "no arg, no env, file exists",
			filename: defaultPath,
			setupEnv: func() {},
			setupFile: func() {
				content := `
tfmigrate {
  migration_dir = "foo"
}
`
				if err := os.WriteFile(defaultPath, []byte(content), 0600); err != nil {
					t.Fatal(err)
				}
			},
			wantDir: "foo",
			wantErr: false,
		},
		{
			desc:      "no arg, no env, no file",
			filename:  defaultPath,
			setupEnv:  func() {},
			setupFile: func() {},
			wantDir:   ".", // Defined in NewDefaultConfig
			wantErr:   false,
		},
		{
			desc:      "load error",
			filename:  "doesnotexist.hcl",
			setupEnv:  func() {},
			setupFile: func() {},
			wantErr:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			// cleanup
			os.Unsetenv("TFMIGRATE_CONFIG")
			os.Remove(defaultPath)

			tc.setupEnv()
			tc.setupFile()

			cfg, err := newConfig(tc.filename)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cfg.MigrationDir != tc.wantDir {
				t.Errorf("MigrationDir = %q; want %q", cfg.MigrationDir, tc.wantDir)
			}
		})
	}
}
