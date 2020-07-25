package tfexec

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
)

// mock functions
type mockFunc func(args ...string) int

func mockEcho(args ...string) int {
	fmt.Println(strings.Join(args[1:], " "))
	return 0
}

func mockFalse(args ...string) int {
	return 1
}

// TestMain customizes initialization of tests for mock.
func TestMain(m *testing.M) {
	mockFunctions := map[string]mockFunc{
		"echo":  mockEcho,
		"false": mockFalse,
	}

	// if test is called with mock mode, behave as a mock.
	if c := os.Getenv("GO_MOCK_COMMAND"); c != "" {
		if f, ok := mockFunctions[c]; ok {
			os.Exit(f(os.Args...))
			return
		}
		// mock function not found.
		fmt.Fprintln(os.Stderr, "mock function not found")
		os.Exit(-1)
		return
	}

	// normal test case
	os.Exit(m.Run())
}

func TestExecutorRun(t *testing.T) {
	cases := []struct {
		desc   string
		args   []string
		stdout string
		stderr string
		ok     bool
	}{
		{
			desc:   "test stdout",
			args:   []string{"echo", "foo", "bar"},
			stdout: "foo bar\n",
			stderr: "",
			ok:     true,
		},
		{
			desc:   "test exit error",
			args:   []string{"false"},
			stdout: "",
			stderr: "",
			ok:     false,
		},
		{
			desc:   "mock function not found",
			args:   []string{"foo"},
			stdout: "",
			stderr: "mock function not found\n",
			ok:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			// call self with GO_MOCK_COMMAND environment variable
			e := NewExecutor(".", []string{"GO_MOCK_COMMAND=" + tc.args[0]})
			cmd, err := e.NewCommandContext(context.Background(), os.Args[0], tc.args[1:]...)
			if err != nil {
				t.Fatalf("failed to NewCommandContext: %s", err)
			}

			err = e.Run(cmd)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			gotStdout := cmd.Stdout()
			gotStderr := cmd.Stderr()
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error, stdout = %s, stderr = %s", gotStdout, gotStderr)
			}
			if gotStdout != tc.stdout {
				t.Errorf("unexpected stdout. got: %s, want: %s", gotStdout, tc.stdout)
			}
			if gotStderr != tc.stderr {
				t.Errorf("unexpected stderr. got: %s, want: %s", gotStderr, tc.stderr)
			}
		})
	}
}

func TestExecutorEnv(t *testing.T) {
	cases := []struct {
		desc        string
		args        []string
		env         []string
		appendKey   string
		appendValue string
		want        string
		ok          bool
	}{
		{
			desc: "test set env with NewExecutor",
			args: []string{"/bin/sh", "-c", "echo $FOO"},
			env:  []string{"FOO=foo"},
			want: "foo\n",
			ok:   true,
		},
		{
			desc:        "test set env with AppendEnv",
			args:        []string{"/bin/sh", "-c", "echo $FOO $BAR"},
			env:         []string{"FOO=foo"},
			appendKey:   "BAR",
			appendValue: "bar",
			want:        "foo bar\n",
			ok:          true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewExecutor(".", tc.env)
			if len(tc.appendKey) > 0 {
				e.AppendEnv(tc.appendKey, tc.appendValue)
			}
			cmd, err := e.NewCommandContext(context.Background(), tc.args[0], tc.args[1:]...)
			if err != nil {
				t.Fatalf("failed to NewCommandContext: %s", err)
			}

			// call real command (not mock).
			// this test may not work with some OS.
			err = e.Run(cmd)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			got := cmd.Stdout()
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error, got = %s", got)
			}
			if got != tc.want {
				t.Errorf("got: %s, want: %s", got, tc.want)
			}
		})
	}
}

func TestExecutorDir(t *testing.T) {
	cases := []struct {
		desc string
		args []string
		dir  string
		want string
		ok   bool
	}{
		{
			desc: "test set dir",
			args: []string{"pwd"},
			dir:  "/bin",
			want: "/bin\n",
			ok:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewExecutor(tc.dir, []string{})
			cmd, err := e.NewCommandContext(context.Background(), tc.args[0], tc.args[1:]...)
			if err != nil {
				t.Fatalf("failed to NewCommandContext: %s", err)
			}

			// call real command (not mock).
			// this test may not work with some OS.
			err = e.Run(cmd)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			got := cmd.Stdout()
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error, got = %s", got)
			}
			if got != tc.want {
				t.Errorf("got: %s, want: %s", got, tc.want)
			}
		})
	}
}
