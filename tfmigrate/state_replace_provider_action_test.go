package tfmigrate

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/minamijoyo/tfmigrate/tfexec"
)

func TestAccStateReplaceProviderActionUsingLegacyTerraform(t *testing.T) {
	tfexec.SkipUnlessAcceptanceTestEnabled(t)

	tfVersion := os.Getenv("TERRAFORM_VERSION")
	if tfVersion != tfexec.LegacyTerraformVersion {
		t.Skipf("skip %s acceptance test for non-legacy Terraform version %s", t.Name(), tfVersion)
	}

	backend := tfexec.GetTestAccBackendS3Config(t.Name())

	source := `
resource "null_resource" "foo" {}
`

	workspace := "default"
	tf := tfexec.SetupTestAccWithApply(t, workspace, backend+source)
	ctx := context.Background()

	actions := []StateAction{
		NewStateReplaceProviderAction("registry.terraform.io/-/null", "registry.terraform.io/hashicorp/null"),
	}

	expected := "replace-provider action requires Terraform version >= 0.13.0"
	m := NewStateMigrator(tf.Dir(), workspace, actions, &MigratorOption{}, false, false)
	err := m.Plan(ctx)
	if err == nil || strings.Contains(err.Error(), expected) {
		t.Fatalf("expected to receive '%s' error using legacy Terraform; got: %s", expected, err)
	}
}

func TestAccStateReplaceProviderAction(t *testing.T) {
	tfexec.SkipUnlessAcceptanceTestEnabled(t)

	tfVersion := os.Getenv("TERRAFORM_VERSION")
	if tfVersion == tfexec.LegacyTerraformVersion {
		t.Skipf("skip %s acceptance test for legacy Terraform version %s", t.Name(), tfVersion)
	}

	backend := tfexec.GetTestAccBackendS3Config(t.Name())

	// To test the use of a non-legacy Terraform CLI version with a legacy
	// Terraform state version, it's necessary to use a state file that was
	// created with a legacy Terraform CLI version, as provided via the
	// test-fixtures/legacy-tfstate directory.
	tfConf, err := os.ReadFile("../test-fixtures/legacy-tfstate/main.tf")
	if err != nil {
		t.Fatalf("error reading test fixture terraform configuration: %s", err)
	}

	source := string(tfConf)

	stateFile, err := os.Open("../test-fixtures/legacy-tfstate/terraform.tfstate")
	if err != nil {
		t.Fatalf("error opening tfstate fixture: %s", err)
	}
	defer stateFile.Close()
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(tfexec.TestS3Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(tfexec.TestS3AccessKey, tfexec.TestS3SecretKey, "")),
	)
	if err != nil {
		t.Fatalf("failed to load aws config: %s", err)
	}
	s3Client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(tfexec.GetTestAccS3Endpoint())
		options.UsePathStyle = true
	})

	// Upload the legacy state file to S3 to pre-seed the backend S3 bucket.
	uploader := manager.NewUploader(s3Client)
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(tfexec.TestS3Bucket),
		Key:    aws.String(tfexec.GetTestAccBackendS3Key(t.Name())),
		Body:   stateFile,
	})
	if err != nil {
		t.Fatalf("failed to upload legacy state file: %s", err)
	}

	workspace := "default"
	tf := tfexec.SetupTestAccForStateReplaceProvider(t, workspace, backend+source)

	registry := "registry.terraform.io"
	tfExecType, _, err := tf.Version(ctx)
	if err != nil {
		t.Fatalf("failed to get tfExecType: %s", err)
	}
	if tfExecType == "opentofu" {
		registry = "registry.opentofu.org"
	}

	actions := []StateAction{
		NewStateReplaceProviderAction(registry+"/-/null", registry+"/hashicorp/null"),
	}

	m := NewStateMigrator(tf.Dir(), workspace, actions, &MigratorOption{}, false, false)
	err = m.Plan(ctx)
	if err != nil {
		t.Fatalf("failed to run migrator plan: %s", err)
	}

	err = m.Apply(ctx)
	if err != nil {
		t.Fatalf("failed to run migrator apply: %s", err)
	}
}
