#!/usr/bin/env bash
set -euo pipefail
# cd ..
# docker build -t seed-generator-plugin:local 


docker run --rm \
  -e CC_STORE_TYPE="${CC_STORE_TYPE:-FS}" \
  -e CC_MANIFEST_ID="${CC_MANIFEST_ID:-manifest}" \
  -e CC_PAYLOAD_ID="${CC_PAYLOAD_ID:-}" \
  -e FSB_ROOT_PATH="${FSB_ROOT_PATH:-/mnt}" \
  -v "$PWD":/mnt \
  seed-generator-plugin:local /app/seed-generator



