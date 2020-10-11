package history

import (
	"reflect"
	"testing"
	"time"
)

func TestHistoryAdd(t *testing.T) {
	cases := []struct {
		desc     string
		h        History
		filename string
		r        Record
		want     History
	}{
		{
			desc: "add a record",
			h: History{
				records: map[string]Record{
					"20201012010101_foo.hcl": Record{
						Type:      "state",
						Name:      "foo",
						AppliedAt: time.Date(2020, 10, 13, 1, 2, 3, 0, time.UTC),
					},
				},
			},
			filename: "20201012020202_foo.hcl",
			r: Record{
				Type:      "state",
				Name:      "bar",
				AppliedAt: time.Date(2020, 10, 13, 4, 5, 6, 0, time.UTC),
			},
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
				},
			},
		},
		{
			desc: "add a duplicated record",
			h: History{
				records: map[string]Record{
					"20201012010101_foo.hcl": Record{
						Type:      "state",
						Name:      "foo",
						AppliedAt: time.Date(2020, 10, 13, 1, 2, 3, 0, time.UTC),
					},
				},
			},
			filename: "20201012010101_foo.hcl",
			r: Record{
				Type:      "state",
				Name:      "bar",
				AppliedAt: time.Date(2020, 10, 13, 4, 5, 6, 0, time.UTC),
			},
			want: History{
				records: map[string]Record{
					"20201012010101_foo.hcl": Record{
						Type:      "state",
						Name:      "bar",
						AppliedAt: time.Date(2020, 10, 13, 4, 5, 6, 0, time.UTC),
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			tc.h.Add(tc.filename, tc.r)
			if !reflect.DeepEqual(tc.h, tc.want) {
				t.Errorf("got: %#v, want: %#v", tc.h, tc.want)
			}
		})
	}
}

func TestHistoryContains(t *testing.T) {
	initialHistory := History{
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
		desc     string
		h        History
		filename string
		want     bool
	}{
		{
			desc:     "exist",
			h:        initialHistory,
			filename: "20201012020202_foo.hcl",
			want:     true,
		},
		{
			desc:     "not exist",
			h:        initialHistory,
			filename: "20201012030303_foo.hcl",
			want:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.h.Contains(tc.filename)
			if got != tc.want {
				t.Errorf("got: %t, want: %t", got, tc.want)
			}
		})
	}
}

func TestHistoryDelete(t *testing.T) {
	cases := []struct {
		desc     string
		h        History
		filename string
		want     History
	}{
		{
			desc: "remove a record",
			h: History{
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
			filename: "20201012020202_foo.hcl",
			want: History{
				records: map[string]Record{
					"20201012010101_foo.hcl": Record{
						Type:      "state",
						Name:      "foo",
						AppliedAt: time.Date(2020, 10, 13, 1, 2, 3, 0, time.UTC),
					},
				},
			},
		},
		{
			desc: "remove non-exist record",
			h: History{
				records: map[string]Record{
					"20201012010101_foo.hcl": Record{
						Type:      "state",
						Name:      "foo",
						AppliedAt: time.Date(2020, 10, 13, 1, 2, 3, 0, time.UTC),
					},
				},
			},
			filename: "20201012030303_foo.hcl",
			want: History{
				records: map[string]Record{
					"20201012010101_foo.hcl": Record{
						Type:      "state",
						Name:      "foo",
						AppliedAt: time.Date(2020, 10, 13, 1, 2, 3, 0, time.UTC),
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			tc.h.Delete(tc.filename)
			if !reflect.DeepEqual(tc.h, tc.want) {
				t.Errorf("got: %#v, want: %#v", tc.h, tc.want)
			}
		})
	}
}

func TestHistoryClear(t *testing.T) {
	cases := []struct {
		desc string
		h    History
		want History
	}{
		{
			desc: "clear all records",
			h: History{
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
			want: History{
				records: map[string]Record{},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			tc.h.Clear()
			if !reflect.DeepEqual(tc.h, tc.want) {
				t.Errorf("got: %#v, want: %#v", tc.h, tc.want)
			}
		})
	}
}
