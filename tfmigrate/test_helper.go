package tfmigrate

import "regexp"

// SwitchBackToRemoteFuncError tests verify error messages, but the
// error message for missing bucket key in the s3 backend differs
// depending on the Terraform version and OpenTofu version.
// Define a helper function to hide the difference.
const testBucketRequiredErrorLegacyTerraform = `Error: "bucket": required field is not set`
const testBucketRequiredErrorTerraform16 = `The attribute "bucket" is required by the backend`
const testBucketRequiredErrorOpenTofu16 = `The "bucket" attribute value must not be empty`

var testBucketRequiredErrorRE = regexp.MustCompile(
	testBucketRequiredErrorLegacyTerraform + `|` +
		testBucketRequiredErrorTerraform16 + `|` +
		testBucketRequiredErrorOpenTofu16)

func containsBucketRequiredError(err error) bool {
	return testBucketRequiredErrorRE.MatchString(err.Error())
}
