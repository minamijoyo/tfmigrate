#!/bin/bash

set -eo pipefail

# TFMIGRATE_EXEC_PATH can be set to tofu or terraform.
TFMIGRATE_EXEC_PATH=${TFMIGRATE_EXEC_PATH:-terraform}

# When using TF_PLUGIN_CACHE_DIR, terraform init is not concurrency safe.
# https://github.com/hashicorp/terraform/issues/25849
# Download the providers and generate a cache once before testing.
mkdir -p "$TF_PLUGIN_CACHE_DIR"

WORK_DIR="./tmp/generate_plugin_cache"
mkdir -p "$WORK_DIR"
pushd "$WORK_DIR"

# Note that it must be written in Terraform v0.12 compatible.
# In other words, you cannot specify a namespace.
cat << EOF > main.tf
terraform {
  required_providers {
    null = "3.2.1"
    time = "0.9.1"
  }
}
EOF

# Starting with Terraform v1.4, the global plugin cache is ignored on the first
# terraform init. This makes caching in CI meaningless. To utilize the cache,
# we use a local filesystem mirror. Strictly speaking, the mirror is only
# available in Terraform v0.13+, but it is hard to compare versions in bash,
# so we use the mirror unless v0.x.
# https://developer.hashicorp.com/terraform/cli/config/config-file#implied-local-mirror-directories
if "$TFMIGRATE_EXEC_PATH" -v | grep 'Terraform v0\.'; then
  echo "skip creating an implied local mirror"
else
  FS_MIRROR="/tmp/plugin-mirror"
  "$TFMIGRATE_EXEC_PATH" providers mirror "${FS_MIRROR}"

  cat << EOF > "$HOME/.terraformrc"
provider_installation {
  filesystem_mirror {
    path    = "/tmp/plugin-mirror"
  }
}
EOF
fi

"$TFMIGRATE_EXEC_PATH" init -input=false -no-color

popd
rm -rf "$WORK_DIR"
