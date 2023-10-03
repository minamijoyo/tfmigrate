package tfmigrate

import "regexp"

// SwitchBackToRemoteFuncError tests verify error messages, but the
// error message for missing bucket key in the s3 backend differs
// depending on the Terraform version.
// Define a helper function to hide the difference.
//
// # Terraform v1.5
//
// ```
// Error: "bucket": required field is not set
// ```
//
// # Terraform v1.6
//
// ```
//
//	Error: Missing Required Value
//
//	  on main.tf line 4, in terraform:
//	   4:   backend "s3" {
//
//	The attribute "bucket" is required by the backend.
//
//	Refer to the backend documentation for additional information which
//	attributes are required.
//
// ```
const testBucketRequiredErrorLegacyTF = `Error: "bucket": required field is not set`
const testBucketRequiredErrorTF16 = `The attribute "bucket" is required by the backend`

var testBucketRequiredErrorRE = regexp.MustCompile(testBucketRequiredErrorLegacyTF + `|` + testBucketRequiredErrorTF16)

func containsBucketRequiredError(err error) bool {
	return testBucketRequiredErrorRE.MatchString(err.Error())
}
