#!/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE_NAME="${IMAGE_NAME:-go-csust-planet}"
SERVER_PATH="${SERVER_PATH:-/home/zhelearn/services/go-csust-planet}"
SSH_PORT="${DEPLOY_PORT:-22}"

usage() {
  cat <<'EOF'
Usage:
  scripts/deploy-remote.sh <target>

Targets:
  main  -> tag latest, service go-csust-planet
  dev   -> tag dev, service go-csust-planet-dev

Environment variables:
  DEPLOY_HOST   Required. SSH host
  DEPLOY_USER   Required. SSH user
  DEPLOY_PORT   Optional. SSH port, defaults to 22
  IMAGE_NAME    Optional. Docker image name, defaults to go-csust-planet
  SERVER_PATH   Optional. Remote deployment path, defaults to /home/zhelearn/services/go-csust-planet
EOF
}

if [[ $# -ne 1 ]]; then
  usage
  exit 1
fi

if [[ -z "${DEPLOY_HOST:-}" || -z "${DEPLOY_USER:-}" ]]; then
  echo "DEPLOY_HOST and DEPLOY_USER are required." >&2
  exit 1
fi

case "$1" in
  main)
    IMAGE_TAG="latest"
    SERVICE_NAME="go-csust-planet"
    ;;
  dev)
    IMAGE_TAG="dev"
    SERVICE_NAME="go-csust-planet-dev"
    ;;
  *)
    echo "Unsupported target: $1" >&2
    usage
    exit 1
    ;;
esac

ARCHIVE_NAME="${IMAGE_NAME}-${IMAGE_TAG}.tar"
LOCAL_ARCHIVE="${ROOT_DIR}/${ARCHIVE_NAME}"
REMOTE_ARCHIVE="${SERVER_PATH}/${ARCHIVE_NAME}"
SSH_TARGET="${DEPLOY_USER}@${DEPLOY_HOST}"

"${ROOT_DIR}/scripts/build-amd64.sh" "${IMAGE_TAG}" "${LOCAL_ARCHIVE}"

cleanup_local() {
  rm -f "${LOCAL_ARCHIVE}"
}

trap cleanup_local EXIT

scp -P "${SSH_PORT}" "${LOCAL_ARCHIVE}" "${SSH_TARGET}:${REMOTE_ARCHIVE}"

ssh -p "${SSH_PORT}" "${SSH_TARGET}" \
  "docker load -i '${REMOTE_ARCHIVE}' && \
   cd '${SERVER_PATH}' && \
   docker compose up -d --no-deps '${SERVICE_NAME}' && \
   rm -f '${REMOTE_ARCHIVE}' && \
   docker image prune -f"
