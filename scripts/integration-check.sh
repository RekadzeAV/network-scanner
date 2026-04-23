#!/usr/bin/env bash
set -euo pipefail

if command -v docker-compose >/dev/null 2>&1; then
  echo "docker-compose detected. Starting integration stack."
  docker-compose up -d
  trap 'docker-compose down' EXIT
else
  echo "docker-compose not found. Running local smoke checks only."
fi

./scripts/smoke-cli-no-topology.sh
./scripts/smoke-cli-topology.sh

echo "Integration check passed."
