#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

echo "== Smoke: CLI with topology =="

go build -o ./release/network-scanner-smoke ./cmd/network-scanner

OUTPUT_FILE="$(mktemp)"
trap 'rm -f "$OUTPUT_FILE"' EXIT

# Keep smoke run fast: single host and small port range.
./release/network-scanner-smoke --network 127.0.0.1/32 --timeout 1 --ports 1-16 --topology >"$OUTPUT_FILE" 2>&1

if ! grep -q "SNMP" "$OUTPUT_FILE"; then
  echo "Smoke failed: expected SNMP summary output in topology mode"
  exit 1
fi

echo "Smoke passed: topology mode prints SNMP summary."
