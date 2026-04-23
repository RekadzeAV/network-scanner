#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

echo "== Smoke: CLI without topology =="

OUTPUT_FILE=""
SMOKE_BIN=""
cleanup() {
  rm -f "${OUTPUT_FILE:-}" "${SMOKE_BIN:-}"
}
trap cleanup EXIT

SMOKE_BIN="$(mktemp "${TMPDIR:-/tmp}/network-scanner-smoke.XXXXXX")"
go build -o "$SMOKE_BIN" ./cmd/network-scanner

OUTPUT_FILE="$(mktemp)"

# Use loopback /32 to keep smoke test fast and deterministic.
"$SMOKE_BIN" --network 127.0.0.1/32 --timeout 1 --ports 1-16 --os-detect-active >"$OUTPUT_FILE" 2>&1

if grep -q "SNMP отчет" "$OUTPUT_FILE"; then
  echo "Smoke failed: SNMP report appears without --topology"
  exit 1
fi

echo "Smoke passed: baseline CLI path works without topology."
