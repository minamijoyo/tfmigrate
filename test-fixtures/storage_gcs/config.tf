terraform {
  # https://www.terraform.io/docs/backends/types/gcs.html
  backend "gcs" {
    bucket = "tfmigrate-gcs"
    prefix = "tfmigrate"
  }
}
