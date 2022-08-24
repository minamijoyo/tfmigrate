terraform {
  # https://www.terraform.io/language/settings/backends/gcs
  backend "gcs" {
    bucket = "tfstate-test"
    prefix = "terraform/state"
  }
}

# https://registry.terraform.io/providers/hashicorp/google/latest/docs
provider "google" {
  project = "dummy"
  region  = "asia-northeast1-a"
}
