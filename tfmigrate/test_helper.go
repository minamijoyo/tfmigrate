package tfmigrate

import "regexp"

// SwitchBackToRemoteFuncError tests verify error messages, but the
// error message for missing bucket key in the s3 backend differs
// depending on the Terraform version.
// Define a helper function to hide the difference.
const testBucketRequiredErrorLegacyTF = `Error: "bucket": required field is not set`
const testBucketRequiredErrorTF16 = `The attribute "bucket" is required by the backend`

var testBucketRequiredErrorRE = regexp.MustCompile(testBucketRequiredErrorLegacyTF + `|` + testBucketRequiredErrorTF16)

func containsBucketRequiredError(err error) bool {
	return testBucketRequiredErrorRE.MatchString(err.Error())
}
