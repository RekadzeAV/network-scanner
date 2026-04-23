#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

echo "== Stage2 P1 closure check (Unix) =="

echo "[1/5] go test ./internal/nettools ./internal/audit ./internal/gui"
go test ./internal/nettools ./internal/audit ./internal/gui

echo "[2/5] go test ./cmd/network-scanner -run Whois"
go test ./cmd/network-scanner -run Whois

echo "[3/5] smoke-cli-tools.sh"
./scripts/smoke-cli-tools.sh

echo "[4/5] build smoke binary"
SMOKE_BIN="$(mktemp "${TMPDIR:-/tmp}/network-scanner-stage2-p1.XXXXXX")"
OUT_FILE="$(mktemp)"
cleanup() {
  rm -f "${SMOKE_BIN:-}" "${OUT_FILE:-}"
}
trap cleanup EXIT
go build -o "$SMOKE_BIN" ./cmd/network-scanner

echo "[5/5] audit-min-severity sanity"
"$SMOKE_BIN" --network 127.0.0.1/32 --ports 1-32 --timeout 1 --threads 1 --audit-open-ports --audit-min-severity critical >"$OUT_FILE" 2>&1
"$SMOKE_BIN" --network 127.0.0.1/32 --ports 1-32 --timeout 1 --threads 1 --audit-open-ports --audit-min-severity high >"$OUT_FILE" 2>&1

echo "Stage2 P1 closure check passed."
