tfmigrate {
  migration_dir = "migrations"
  
  # Optional: Default executable path for all operations
  # This is equivalent to setting TFMIGRATE_EXEC_PATH environment variable
  exec_path = "terraform"
  
  # Optional: Use a different executable for source operations in multi-state migrations
  # For example, to use Terraform for source state and OpenTofu for destination state
  from_tf_exec_path = "terraform"
  to_tf_exec_path = "tofu"
  
  # Optional: Configure history tracking
  history {
    storage "local" {
      path = ".terraform/tfmigrate.json" 
    }
  }
}
