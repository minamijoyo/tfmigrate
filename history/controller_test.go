package history

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestLoadMigrationFileNames(t *testing.T) {
	cases := []struct {
		desc  string
		files []string
		want  []string
		ok    bool
	}{
		{
			desc: "hcl",
			files: []string{
				"20201012010101_foo.hcl",
				"20201012020202_foo.hcl",
			},
			want: []string{
				"20201012010101_foo.hcl",
				"20201012020202_foo.hcl",
			},
			ok: true,
		},
		{
			desc: "json",
			files: []string{
				"20201012010101_foo.hcl",
				"20201012020202_foo.json",
				"20201012020202_foo.txt",
			},
			want: []string{
				"20201012010101_foo.hcl",
				"20201012020202_foo.json",
			},
			ok: true,
		},
		{
			desc: "ignore hidden files",
			files: []string{
				"20201012010101_foo.hcl",
				"20201012020202_foo.json",
				".tfmigrate.hcl",
				".terraform.lock.hcl",
			},
			want: []string{
				"20201012010101_foo.hcl",
				"20201012020202_foo.json",
			},
			ok: true,
		},
		{
			desc: "unsorted",
			files: []string{
				"20201012020202_foo.hcl",
				"20201012010101_foo.hcl",
			},
			want: []string{
				"20201012010101_foo.hcl",
				"20201012020202_foo.hcl",
			},
			ok: true,
		},
		{
			desc:  "empty",
			files: []string{},
			want:  []string{},
			ok:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			migrationDir, err := ioutil.TempDir("", "migrationDir")
			if err != nil {
				t.Fatalf("failed to craete temp dir: %s", err)
			}
			t.Cleanup(func() { os.RemoveAll(migrationDir) })

			for _, filename := range tc.files {
				err = ioutil.WriteFile(filepath.Join(migrationDir, filename), []byte{}, 0600)
				if err != nil {
					t.Fatalf("failed to write dummy migration file: %s", err)
				}
			}

			got, err := loadMigrationFileNames(migrationDir)

			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %#v", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}

			if tc.ok {
				if diff := cmp.Diff(got, tc.want); diff != "" {
					t.Errorf("got = %#v, want = %#v, diff = %s", got, tc.want, diff)
				}
			}
		})
	}
}

func TestLoadHistory(t *testing.T) {
	cases := []struct {
		desc   string
		config StorageConfig
		want   *History
		ok     bool
	}{
		{
			desc: "simple",
			config: &MockStorageConfig{
				Data: `{
    "version": 1,
    "records": {
        "20201012010101_foo.hcl": {
            "type": "state",
            "name": "foo",
            "applied_at": "2020-10-13T01:02:03Z"
        },
        "20201012020202_foo.hcl": {
            "type": "state",
            "name": "bar",
            "applied_at": "2020-10-13T04:05:06Z"
        }
    }
}`,
				WriteError: false,
				ReadError:  false,
			},
			want: &History{
				records: map[string]Record{
					"20201012010101_foo.hcl": Record{
						Type:      "state",
						Name:      "foo",
						AppliedAt: time.Date(2020, 10, 13, 1, 2, 3, 0, time.UTC),
					},
					"20201012020202_foo.hcl": Record{
						Type:      "state",
						Name:      "bar",
						AppliedAt: time.Date(2020, 10, 13, 4, 5, 6, 0, time.UTC),
					},
				},
			},
			ok: true,
		},
		{
			desc: "empty",
			config: &MockStorageConfig{
				Data:       "",
				WriteError: false,
				ReadError:  false,
			},
			want: newEmptyHistory(),
			ok:   true,
		},
		{
			desc: "read error",
			config: &MockStorageConfig{
				Data:       "",
				WriteError: false,
				ReadError:  true,
			},
			want: nil,
			ok:   false,
		},
		{
			desc: "invalid format",
			config: &MockStorageConfig{
				Data:       "foo",
				WriteError: false,
				ReadError:  false,
			},
			want: nil,
			ok:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := loadHistory(context.Background(), tc.config)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %#v", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}

			if tc.ok {
				if diff := cmp.Diff(*got, *tc.want, cmp.AllowUnexported(*got)); diff != "" {
					t.Errorf("got = %#v, want = %#v, diff = %s", got, tc.want, diff)
				}
			}
		})
	}
}

func TestControllerSave(t *testing.T) {
	cases := []struct {
		desc   string
		config *MockStorageConfig
		h      *History
		want   []byte
		ok     bool
	}{
		{
			desc: "simple",
			config: &MockStorageConfig{
				Data:       "",
				WriteError: false,
				ReadError:  false,
			},
			h: newEmptyHistory(),
			want: []byte(`{
    "version": 1,
    "records": {}
}`),
			ok: true,
		},
		{
			desc: "write error",
			config: &MockStorageConfig{
				Data:       "",
				WriteError: true,
				ReadError:  false,
			},
			h: newEmptyHistory(),
			want: []byte(`{
    "version": 1,
    "records": {}
}`),
			ok: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			c := &Controller{
				history: *tc.h,
				config: Config{
					Storage: tc.config,
				},
			}
			err := c.Save(context.Background())
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}

			if tc.ok {
				got := []byte(tc.config.s.data)
				if string(got) != string(tc.want) {
					t.Errorf("got: %s, want: %s", string(got), string(tc.want))
				}
			}
		})
	}
}

func TestUnappliedMigrations(t *testing.T) {
	cases := []struct {
		desc       string
		migrations []string
		history    History
		want       []string
	}{
		{
			desc: "simple",
			migrations: []string{
				"20201012010101_foo.hcl",
				"20201012020202_foo.hcl",
				"20201012030303_foo.hcl",
				"20201012040404_foo.hcl",
			},
			history: History{
				records: map[string]Record{
					"20201012010101_foo.hcl": Record{
						Type:      "state",
						Name:      "foo",
						AppliedAt: time.Date(2020, 10, 13, 1, 2, 3, 0, time.UTC),
					},
					"20201012020202_foo.hcl": Record{
						Type:      "state",
						Name:      "bar",
						AppliedAt: time.Date(2020, 10, 13, 4, 5, 6, 0, time.UTC),
					},
				},
			},
			want: []string{
				"20201012030303_foo.hcl",
				"20201012040404_foo.hcl",
			},
		},
		{
			desc: "all applied",
			migrations: []string{
				"20201012010101_foo.hcl",
				"20201012020202_foo.hcl",
			},
			history: History{
				records: map[string]Record{
					"20201012010101_foo.hcl": Record{
						Type:      "state",
						Name:      "foo",
						AppliedAt: time.Date(2020, 10, 13, 1, 2, 3, 0, time.UTC),
					},
					"20201012020202_foo.hcl": Record{
						Type:      "state",
						Name:      "bar",
						AppliedAt: time.Date(2020, 10, 13, 4, 5, 6, 0, time.UTC),
					},
				},
			},
			want: []string{},
		},
		{
			desc: "ignore a missing migration file include in history",
			migrations: []string{
				"20201012020202_foo.hcl",
				"20201012030303_foo.hcl",
				"20201012040404_foo.hcl",
			},
			history: History{
				records: map[string]Record{
					"20201012010101_foo.hcl": Record{
						Type:      "state",
						Name:      "foo",
						AppliedAt: time.Date(2020, 10, 13, 1, 2, 3, 0, time.UTC),
					},
					"20201012020202_foo.hcl": Record{
						Type:      "state",
						Name:      "bar",
						AppliedAt: time.Date(2020, 10, 13, 4, 5, 6, 0, time.UTC),
					},
				},
			},
			want: []string{
				"20201012030303_foo.hcl",
				"20201012040404_foo.hcl",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			c := &Controller{
				migrations: tc.migrations,
				history:    tc.history,
			}

			got := c.UnappliedMigrations()
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got = %#v, want = %#v", got, tc.want)
			}
		})
	}
}

func TestControllerHistoryLength(t *testing.T) {
	cases := []struct {
		desc       string
		migrations []string
		history    History
		want       int
	}{
		{
			desc: "simple",
			migrations: []string{
				"20201012010101_foo.hcl",
				"20201012020202_foo.hcl",
				"20201012030303_foo.hcl",
				"20201012040404_foo.hcl",
			},
			history: History{
				records: map[string]Record{
					"20201012010101_foo.hcl": Record{
						Type:      "state",
						Name:      "foo",
						AppliedAt: time.Date(2020, 10, 13, 1, 2, 3, 0, time.UTC),
					},
					"20201012020202_foo.hcl": Record{
						Type:      "state",
						Name:      "bar",
						AppliedAt: time.Date(2020, 10, 13, 4, 5, 6, 0, time.UTC),
					},
				},
			},
			want: 2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			c := &Controller{
				migrations: tc.migrations,
				history:    tc.history,
			}

			got := c.HistoryLength()
			if got != tc.want {
				t.Errorf("got = %d, want = %d", got, tc.want)
			}
		})
	}
}

func TestControllerAlreadApplied(t *testing.T) {
	migrations := []string{
		"20201012010101_foo.hcl",
		"20201012020202_foo.hcl",
		"20201012030303_foo.hcl",
		"20201012040404_foo.hcl",
	}
	history := History{
		records: map[string]Record{
			"20201012010101_foo.hcl": Record{
				Type:      "state",
				Name:      "foo",
				AppliedAt: time.Date(2020, 10, 13, 1, 2, 3, 0, time.UTC),
			},
			"20201012020202_foo.hcl": Record{
				Type:      "state",
				Name:      "bar",
				AppliedAt: time.Date(2020, 10, 13, 4, 5, 6, 0, time.UTC),
			},
		},
	}
	cases := []struct {
		desc       string
		migrations []string
		history    History
		filename   string
		want       bool
	}{
		{
			desc:       "unapplied",
			migrations: migrations,
			history:    history,
			filename:   "20201012030303_foo.hcl",
			want:       false,
		},
		{
			desc:       "applied",
			migrations: migrations,
			history:    history,
			filename:   "20201012020202_foo.hcl",
			want:       true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			c := &Controller{
				migrations: tc.migrations,
				history:    tc.history,
			}

			got := c.AlreadyApplied(tc.filename)
			if got != tc.want {
				t.Errorf("got = %t, want = %t", got, tc.want)
			}
		})
	}
}

func TestControllerAddRecord(t *testing.T) {
	migrations := []string{
		"20201012010101_foo.hcl",
		"20201012020202_foo.hcl",
		"20201012030303_foo.hcl",
		"20201012040404_foo.hcl",
	}
	history := History{
		records: map[string]Record{
			"20201012010101_foo.hcl": Record{
				Type:      "state",
				Name:      "foo",
				AppliedAt: time.Date(2020, 10, 13, 1, 2, 3, 0, time.UTC),
			},
			"20201012020202_foo.hcl": Record{
				Type:      "state",
				Name:      "bar",
				AppliedAt: time.Date(2020, 10, 13, 4, 5, 6, 0, time.UTC),
			},
		},
	}
	cases := []struct {
		desc          string
		migrations    []string
		history       History
		filename      string
		migrationType string
		name          string
		appliedAt     time.Time
		want          History
	}{
		{
			desc:          "add",
			migrations:    migrations,
			history:       history,
			filename:      "20201012030303_foo.hcl",
			migrationType: "multi_state",
			name:          "baz",
			appliedAt:     time.Date(2020, 10, 13, 7, 8, 9, 0, time.UTC),
			want: History{
				records: map[string]Record{
					"20201012010101_foo.hcl": Record{
						Type:      "state",
						Name:      "foo",
						AppliedAt: time.Date(2020, 10, 13, 1, 2, 3, 0, time.UTC),
					},
					"20201012020202_foo.hcl": Record{
						Type:      "state",
						Name:      "bar",
						AppliedAt: time.Date(2020, 10, 13, 4, 5, 6, 0, time.UTC),
					},
					"20201012030303_foo.hcl": Record{
						Type:      "multi_state",
						Name:      "baz",
						AppliedAt: time.Date(2020, 10, 13, 7, 8, 9, 0, time.UTC),
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			c := &Controller{
				migrations: tc.migrations,
				history:    tc.history,
			}

			c.AddRecord(tc.filename, tc.migrationType, tc.name, &tc.appliedAt)
			got := tc.history
			if diff := cmp.Diff(got, tc.want, cmp.AllowUnexported(got)); diff != "" {
				t.Errorf("got = %#v, want = %#v, diff = %s", got, tc.want, diff)
			}
		})
	}
}
