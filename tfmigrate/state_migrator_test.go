package tfmigrate

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

func TestAccStateMigratorApply(t *testing.T) {
	tfexec.SkipUnlessAcceptanceTestEnabled(t)

	backend := `
terraform {
  # https://www.terraform.io/docs/backends/types/s3.html
  backend "s3" {
    region = "ap-northeast-1"
    bucket = "tfstate-test"
    key    = "test/terraform.tfstate"

    // mock s3 endpoint with localstack
    endpoint                    = "http://localstack:4566"
    access_key                  = "dummy"
    secret_key                  = "dummy"
    skip_credentials_validation = true
    skip_metadata_api_check     = true
    force_path_style            = true
  }
}

# https://www.terraform.io/docs/providers/aws/index.html
# https://www.terraform.io/docs/providers/aws/guides/custom-service-endpoints.html#localstack
provider "aws" {
  region = "ap-northeast-1"

  access_key                  = "dummy"
  secret_key                  = "dummy"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_region_validation      = true
  skip_requesting_account_id  = true
  s3_force_path_style         = true

  // mock endpoints with localstack
  endpoints {
    s3  = "http://localstack:4566"
    ec2 = "http://localstack:4566"
  }
}
`

	source := `
resource "aws_security_group" "foo" {}
resource "aws_security_group" "bar" {}
resource "aws_security_group" "baz" {}
`
	e := tfexec.SetupTestAcc(t, source+backend)
	tf := tfexec.NewTerraformCLI(e)
	ctx := context.Background()

	err := tf.Init(ctx, "", "-input=false", "-no-color")
	if err != nil {
		t.Fatalf("failed to run terraform init: %s", err)
	}

	err = tf.Apply(ctx, nil, "", "-input=false", "-no-color", "-auto-approve")
	if err != nil {
		t.Fatalf("failed to run terraform apply: %s", err)
	}

	updatedSource := `
resource "aws_security_group" "foo2" {}
resource "aws_security_group" "bar2" {}
resource "aws_security_group" "baz" {}
`

	if err := ioutil.WriteFile(filepath.Join(e.Dir(), "main.tf"), []byte(updatedSource+backend), 0644); err != nil {
		t.Fatalf("failed to update source: %s", err)
	}

	_, err = tf.Plan(ctx, nil, "", "-input=false", "-no-color", "-detailed-exitcode")
	if err != nil {
		if exitErr, ok := err.(tfexec.ExitError); ok {
			if exitErr.ExitCode() != 2 {
				t.Fatalf("failed to run terraform plan before migrate (expected diff): %s", err)
			}
		} else {
			t.Fatalf("failed to run terraform plan before migrate (unexpected error): %s", err)
		}
	} else {
		t.Fatalf("terraform plan success but expected diff before migrate: %s", err)
	}

	actions := []StateAction{
		NewStateMvAction("aws_security_group.foo", "aws_security_group.foo2"),
		NewStateMvAction("aws_security_group.bar", "aws_security_group.bar2"),
	}

	m := NewStateMigrator(e.Dir(), actions, nil)
	err = m.Plan(context.Background())
	if err != nil {
		t.Fatalf("failed to run migrator plan: %s", err)
	}

	err = m.Apply(context.Background())
	if err != nil {
		t.Fatalf("failed to run migrator apply: %s", err)
	}

	got, err := tf.StateList(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("failed to run terraform state list: %s", err)
	}

	want := []string{
		"aws_security_group.foo2",
		"aws_security_group.bar2",
		"aws_security_group.baz",
	}
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got state: %v, want state: %v", got, want)
	}

}
