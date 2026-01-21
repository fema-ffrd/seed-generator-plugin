#!/usr/bin/env bash
set -euo pipefail
cd .. #Assumes you are calling the script from /uatest
docker build . -t seed-generator-plugin:local
cd uatest

# Run Block Generator First
cp payloads/payload.blocks.json payload
docker run --rm \
  -e CC_STORE_TYPE="${CC_STORE_TYPE:-FS}" \
  -e CC_MANIFEST_ID="${CC_MANIFEST_ID:-manifest}" \
  -e CC_PAYLOAD_ID="${CC_PAYLOAD_ID:-}" \
  -e FSB_ROOT_PATH="${FSB_ROOT_PATH:-/mnt}" \
  -v "$PWD":/mnt \
  seed-generator-plugin:local /app/seed-generator
rm payload

# Run Seed Generator Second
cp payloads/payload.seeds.json payload
docker run --rm \
  -e CC_STORE_TYPE="${CC_STORE_TYPE:-FS}" \
  -e CC_MANIFEST_ID="${CC_MANIFEST_ID:-manifest}" \
  -e CC_PAYLOAD_ID="${CC_PAYLOAD_ID:-}" \
  -e FSB_ROOT_PATH="${FSB_ROOT_PATH:-/mnt}" \
  -v "$PWD":/mnt \
  seed-generator-plugin:local /app/seed-generator

