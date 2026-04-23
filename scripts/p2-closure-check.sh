#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

echo "== P2 closure check (Unix) =="

echo "[1/5] go test ./..."
go test ./...

echo "[2/5] smoke-cli-no-topology.sh"
./scripts/smoke-cli-no-topology.sh

echo "[3/5] smoke-cli-topology.sh"
./scripts/smoke-cli-topology.sh

echo "[4/5] smoke-cli-tools.sh"
./scripts/smoke-cli-tools.sh

echo "[5/5] focused P2 flags sanity"
SMOKE_BIN="$(mktemp "${TMPDIR:-/tmp}/network-scanner-smoke.XXXXXX")"
OUT_FILE="$(mktemp)"
cleanup() {
  rm -f "${SMOKE_BIN:-}" "${OUT_FILE:-}"
}
trap cleanup EXIT

go build -o "$SMOKE_BIN" ./cmd/network-scanner
"$SMOKE_BIN" --network 127.0.0.1/32 --ports 80,443 --timeout 1 --grab-banners --show-raw-banners --os-detect-active >"$OUT_FILE" 2>&1

# Invalid MAC must fail in WOL mode.
if "$SMOKE_BIN" --wol-mac invalid-mac >/dev/null 2>&1; then
  echo "P2 closure failed: invalid WOL MAC should return non-zero exit code"
  exit 1
fi

echo "P2 closure check passed."

