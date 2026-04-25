#!/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE_NAME="${IMAGE_NAME:-go-csust-planet}"
IMAGE_TAG="${1:-latest}"
OUTPUT_TAR="${2:-${IMAGE_NAME}-${IMAGE_TAG}.tar}"

cd "$ROOT_DIR"

docker buildx build \
  --platform linux/amd64 \
  -t "${IMAGE_NAME}:${IMAGE_TAG}" \
  -o "type=docker,dest=${OUTPUT_TAR}" \
  .
