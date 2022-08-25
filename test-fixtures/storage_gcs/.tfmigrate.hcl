tfmigrate {
  migration_dir = "./tfmigrate"
  history {
    storage "gcs" {
      bucket = "tfstate-test"
      name   = "tfmigrate/history.json"
    }
  }
}
