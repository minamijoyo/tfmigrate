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
  s3_use_path_style           = true

  // mock endpoints with localstack
  endpoints {
    s3  = "http://localstack:4566"
    ec2 = "http://localstack:4566"
    iam = "http://localstack:4566"
  }
}
