tfmigrate {
  migration_dir = "./tfmigrate"
  history {
    storage "s3" {
      bucket = "tfstate-test"
      key    = "tfmigrate/history.json"
      region = "ap-northeast-1"

      endpoint                    = "http://localstack:4566"
      access_key                  = "dummy"
      secret_key                  = "dummy"
      skip_credentials_validation = true
      skip_metadata_api_check     = true
      force_path_style            = true
    }
  }
}
