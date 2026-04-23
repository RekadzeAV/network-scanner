#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

echo "== P1 closure check (Unix) =="

echo "[1/4] go test ./..."
go test ./...

echo "[2/4] smoke-cli-no-topology.sh"
./scripts/smoke-cli-no-topology.sh

echo "[3/4] smoke-cli-topology.sh"
./scripts/smoke-cli-topology.sh

echo "[4/4] smoke-cli-tools.sh"
./scripts/smoke-cli-tools.sh

echo "P1 closure check passed."
