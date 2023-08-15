package s3

import storage "github.com/minamijoyo/tfmigrate-storage"

// Config is a config for s3 storage.
// This is expected to have almost the same options as Terraform s3 backend.
// https://www.terraform.io/docs/backends/types/s3.html
// However, it has many minor options and it's a pain to test all options from
// first, so we added only options we need for now.
type Config struct {
	// Name of the bucket.
	Bucket string `hcl:"bucket"`
	// Path to the migration history file.
	Key string `hcl:"key"`

	// AWS region.
	Region string `hcl:"region,optional"`
	// Custom endpoint for the AWS S3 API.
	Endpoint string `hcl:"endpoint,optional"`
	// AWS access key.
	AccessKey string `hcl:"access_key,optional"`
	// AWS secret key.
	SecretKey string `hcl:"secret_key,optional"`
	// Name of AWS profile in AWS shared credentials file.
	Profile string `hcl:"profile,optional"`
	// Amazon Resource Name (ARN) of the IAM Role to assume.
	RoleARN string `hcl:"role_arn,optional"`
	// Skip credentials validation via the STS API.
	SkipCredentialsValidation bool `hcl:"skip_credentials_validation,optional"`
	// Skip usage of EC2 Metadata API.
	SkipMetadataAPICheck bool `hcl:"skip_metadata_api_check,optional"`
	// Enable path-style S3 URLs (https://<HOST>/<BUCKET>
	// instead of https://<BUCKET>.<HOST>).
	ForcePathStyle bool `hcl:"force_path_style,optional"`
	// SSE KMS Key Id for optional server-side encryption enablement
	KmsKeyID string `hcl:"kms_key_id,optional"`
}

// Config implements a storage.Config.
var _ storage.Config = (*Config)(nil)

// NewStorage returns a new instance of storage.Storage.
func (c *Config) NewStorage() (storage.Storage, error) {
	return NewStorage(c, nil)
}
