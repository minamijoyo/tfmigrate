#!/bin/bash

set -eo pipefail

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

terraform init -input=false -no-color

popd
rm -rf "$WORK_DIR"
