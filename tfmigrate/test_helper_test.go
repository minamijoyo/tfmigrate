package tfmigrate

import (
	"errors"
	"testing"
)

func TestContainsBucketRequiredError(t *testing.T) {
	cases := []struct {
		desc string
		msg  string
		want bool
	}{
		{
			desc: "terraform v1.5",
			msg:  `Error: "bucket": required field is not set`,
			want: true,
		},
		{
			desc: "terraform v1.6",
			msg: `
Error: Missing Required Value

  on main.tf line 4, in terraform:
   4:   backend "s3" {

The attribute "bucket" is required by the backend.

Refer to the backend documentation for additional information which
attributes are required.

`,
			want: true,
		},
		{
			desc: "unknown",
			msg:  `Error: unknown`,
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got := containsBucketRequiredError(errors.New(tc.msg))
			if got != tc.want {
				t.Errorf("got: %t, want: %t", got, tc.want)
			}
		})
	}
}
