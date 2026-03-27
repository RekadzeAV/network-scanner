#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

echo "== Smoke: CLI without topology =="

go build -o ./release/network-scanner-smoke ./cmd/network-scanner

OUTPUT_FILE="$(mktemp)"
trap 'rm -f "$OUTPUT_FILE"' EXIT

# Use loopback /32 to keep smoke test fast and deterministic.
./release/network-scanner-smoke --network 127.0.0.1/32 --timeout 1 --ports 1-16 >"$OUTPUT_FILE" 2>&1

if grep -q "SNMP отчет" "$OUTPUT_FILE"; then
  echo "Smoke failed: SNMP report appears without --topology"
  exit 1
fi

echo "Smoke passed: baseline CLI path works without topology."
