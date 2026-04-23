#!/usr/bin/env bash
set -euo pipefail

echo "==> Checking Go toolchain"
if ! command -v go >/dev/null 2>&1; then
  echo "Go is not installed. Install Go 1.24+ first."
  exit 1
fi

echo "==> Go version"
go version

echo "==> Downloading module dependencies"
go mod download

echo "==> Building CLI binary"
go build -o network-scanner ./cmd/network-scanner

echo "==> Running unit/integration tests"
go test ./...

echo "Bootstrap completed successfully."
