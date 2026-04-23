#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

echo "== Stage2 P3 closure check (Unix) =="

echo "[1/5] go test ./internal/cve ./internal/report ./internal/remoteexec ./cmd/network-scanner"
go test ./internal/cve ./internal/report ./internal/remoteexec ./cmd/network-scanner

echo "[2/5] build smoke binary"
SMOKE_BIN="$(mktemp "${TMPDIR:-/tmp}/network-scanner-stage2-p3.XXXXXX")"
OUT_FILE="$(mktemp)"
AUDIT_FILE="$(mktemp)"
cleanup() {
  rm -f "${SMOKE_BIN:-}" "${OUT_FILE:-}" "${AUDIT_FILE:-}"
}
trap cleanup EXIT
go build -o "$SMOKE_BIN" ./cmd/network-scanner

echo "[3/5] security-report auto(redacted) sanity"
"$SMOKE_BIN" --network 127.0.0.1/32 --ports 22,80 --timeout 1 --threads 1 --risk-signatures --security-report-file auto --security-report-redact=true >"$OUT_FILE" 2>&1
if ! rg -q "security-report-redacted-.*\\.html" "$OUT_FILE"; then
  echo "Expected redacted auto report filename in output"
  exit 1
fi

echo "[4/5] security-report unredacted consent sanity"
"$SMOKE_BIN" --network 127.0.0.1/32 --ports 22,80 --timeout 1 --threads 1 --risk-signatures --security-report-file auto --security-report-redact=false --security-report-unsafe-consent I_UNDERSTAND_UNREDACTED_REPORT >"$OUT_FILE" 2>&1
if ! rg -q "security-report-unredacted-.*\\.html" "$OUT_FILE"; then
  echo "Expected unredacted auto report filename in output"
  exit 1
fi
if ! rg -q "report-id=" "$OUT_FILE"; then
  echo "Expected report-id marker in output"
  exit 1
fi

echo "[5/5] remote-exec dry-run strict policy sanity"
"$SMOKE_BIN" --remote-exec-transport ssh --remote-exec-target 127.0.0.1 --remote-exec-command "hostname" --remote-exec-policy-file config/remote-exec-policy.example.json --remote-exec-policy-strict --remote-exec-consent I_UNDERSTAND --remote-exec-dry-run=true --remote-exec-audit-log "$AUDIT_FILE" >"$OUT_FILE" 2>&1 || true
# Using example policy may not include 127.0.0.1 and should be blocked by allowlist with explicit policy error.
if ! rg -q "Remote exec policy ошибка|target is not in allowlist" "$OUT_FILE"; then
  echo "Expected policy/allowlist guardrail output"
  exit 1
fi

echo "Stage2 P3 closure check passed."
