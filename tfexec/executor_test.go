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
		os.Exit(-1)
		return
	}

	// normal test case
	os.Exit(m.Run())
}

func TestExecutorRun(t *testing.T) {
	cases := []struct {
		desc string
		args []string
		want string
		ok   bool
	}{
		{
			desc: "test stdout",
			args: []string{"echo", "foo", "bar"},
			want: "foo bar\n",
			ok:   true,
		},
		{
			desc: "test exit error",
			args: []string{"false"},
			want: "",
			ok:   false,
		},
		{
			desc: "mock function not found",
			args: []string{"foo"},
			want: "",
			ok:   false,
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

			err = cmd.Run()
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

func TestExecutorEnv(t *testing.T) {
	cases := []struct {
		desc string
		args []string
		env  []string
		want string
		ok   bool
	}{
		{
			desc: "test set env",
			args: []string{"/bin/sh", "-c", "echo $FOO"},
			env:  []string{"FOO=foo"},
			want: "foo\n",
			ok:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewExecutor(".", tc.env)
			cmd, err := e.NewCommandContext(context.Background(), tc.args[0], tc.args[1:]...)
			if err != nil {
				t.Fatalf("failed to NewCommandContext: %s", err)
			}

			// call real command (not mock).
			// this test may not work with some OS.
			err = cmd.Run()
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
			err = cmd.Run()
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
