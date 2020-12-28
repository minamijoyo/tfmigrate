#!/bin/bash
set -e

# Create a plugin cache directory in advance.
# The terraform command doesn't create it automatically.
if [ -n "${TF_PLUGIN_CACHE_DIR}" ]; then
  mkdir -p "${TF_PLUGIN_CACHE_DIR}"
fi

exec "$@"
