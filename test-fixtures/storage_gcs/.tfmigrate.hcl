tfmigrate {
  migration_dir = "./tfmigrate"
  history {
    storage "gcs" {
      bucket = "tfmigrate-gcs"
      prefix    = "tfmigrate/history.json"
    }
  }
}
