package tfexec

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// setupTestAcc is a common setup helper for acceptance tests.
func setupTestAcc(t *testing.T, source string) Executor {
	workDir, err := setupTestWorkDir(source)
	if err != nil {
		t.Fatalf("failed to setup work dir: %s", err)
	}
	t.Cleanup(func() { os.RemoveAll(workDir) })

	e := NewExecutor(workDir, os.Environ())
	if err := setupTestPluginCacheDir(e); err != nil {
		t.Fatalf("failed to set plugin cache dir: %s", err)
	}

	return e
}

// isAcceptanceTestEnabled returns true if acceptance tests should be run.
func isAcceptanceTestEnabled() bool {
	return os.Getenv("TEST_ACC") == "1"
}

// setupTestWorkDir creates temporary working directory with a given source for testing.
func setupTestWorkDir(source string) (string, error) {
	workDir, err := ioutil.TempDir("", "workDir")
	if err != nil {
		return "", fmt.Errorf("failed to create work dir: %s", err)
	}

	if err := ioutil.WriteFile(filepath.Join(workDir, "main.tf"), []byte(source), 0644); err != nil {
		os.RemoveAll(workDir)
		return "", fmt.Errorf("failed to create main.tf: %s", err)
	}
	return workDir, nil
}

// setupTestPluginCacheDir sets TF_PLUGIN_CACHE_DIR to a given executor.
func setupTestPluginCacheDir(e Executor) error {
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current dir: %s", err)
	}
	e.AppendEnv("TF_PLUGIN_CACHE_DIR", filepath.Join(pwd, "tmp/plugin-cache"))
	return nil
}

func TestTerraformCLIRun(t *testing.T) {
	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		args         []string
		execPath     string
		want         string
		ok           bool
	}{
		{
			desc: "run terraform version",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "version"},
					stdout:   "Terraform v0.12.28\n",
					exitCode: 0,
				},
			},
			args: []string{"version"},
			want: "Terraform v0.12.28\n",
			ok:   true,
		},
		{
			desc: "failed to run terraform version",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "version"},
					exitCode: 1,
				},
			},
			args: []string{"version"},
			want: "",
			ok:   false,
		},
		{
			desc: "with execPath (no space)",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform-0.12.28", "version"},
					stdout:   "Terraform v0.12.28\n",
					exitCode: 0,
				},
			},
			args:     []string{"version"},
			execPath: "terraform-0.12.28",
			want:     "Terraform v0.12.28\n",
			ok:       true,
		},
		{
			desc: "with execPath (spaces)",
			mockCommands: []*mockCommand{
				{
					args:     []string{"direnv", "exec", ".", "terraform", "version"},
					stdout:   "Terraform v0.12.28\n",
					exitCode: 0,
				},
			},
			args:     []string{"version"},
			execPath: "direnv exec . terraform",
			want:     "Terraform v0.12.28\n",
			ok:       true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			terraformCLI.SetExecPath(tc.execPath)
			got, _, err := terraformCLI.Run(context.Background(), tc.args...)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error, got = %s", got)
			}
			if got != tc.want {
				t.Errorf("got: %s, want: %s", got, tc.want)
			}
		})
	}
}
