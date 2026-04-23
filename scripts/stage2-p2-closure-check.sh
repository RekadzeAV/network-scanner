#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

echo "== Stage2 P2 closure check (Unix) =="

echo "[1/5] go test ./internal/risksignature ./internal/devicecontrol ./internal/report ./internal/gui"
go test ./internal/risksignature ./internal/devicecontrol ./internal/report ./internal/gui

echo "[2/5] smoke-cli-tools.sh"
./scripts/smoke-cli-tools.sh

echo "[3/5] build smoke binary"
SMOKE_BIN="$(mktemp "${TMPDIR:-/tmp}/network-scanner-stage2-p2.XXXXXX")"
REPORT_FILE="$(mktemp "${TMPDIR:-/tmp}/security-report-stage2-p2.XXXXXX.html")"
cleanup() {
  rm -f "${SMOKE_BIN:-}" "${REPORT_FILE:-}"
}
trap cleanup EXIT
go build -o "$SMOKE_BIN" ./cmd/network-scanner

echo "[4/5] risk-signatures + security-report sanity"
"$SMOKE_BIN" --network 127.0.0.1/32 --ports 22,80 --timeout 1 --threads 1 --risk-signatures --security-report-file "$REPORT_FILE" > /dev/null 2>&1
rg -q "CVE Findings" "$REPORT_FILE"
rg -q "Risk Signature Findings" "$REPORT_FILE"

echo "[5/5] device-control negative cases"
if "$SMOKE_BIN" --device-action reboot --device-target http://127.0.0.1 --device-vendor generic-http > /dev/null 2>&1; then
  echo "Expected reboot without --device-confirm to fail"
  exit 1
fi
if "$SMOKE_BIN" --device-action status > /dev/null 2>&1; then
  echo "Expected status without --device-target to fail"
  exit 1
fi

echo "Stage2 P2 closure check passed."
