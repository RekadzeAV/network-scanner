#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

echo "== Smoke: CLI tools mode (ping/dns) =="

PING_OUTPUT_FILE=""
DNS_OUTPUT_FILE=""
AUDIT_OUTPUT_FILE=""
AUDIT_HIGH_OUTPUT_FILE=""
SMOKE_BIN=""
cleanup() {
  rm -f "${PING_OUTPUT_FILE:-}" "${DNS_OUTPUT_FILE:-}" "${AUDIT_OUTPUT_FILE:-}" "${AUDIT_HIGH_OUTPUT_FILE:-}" "${SMOKE_BIN:-}"
}
trap cleanup EXIT

SMOKE_BIN="$(mktemp "${TMPDIR:-/tmp}/network-scanner-smoke.XXXXXX")"
go build -o "$SMOKE_BIN" ./cmd/network-scanner

# Deterministic whois CLI flow check without external network.
go test ./cmd/network-scanner -run WhoisUsesRDAPFallback -count=1

PING_OUTPUT_FILE="$(mktemp)"
DNS_OUTPUT_FILE="$(mktemp)"
AUDIT_OUTPUT_FILE="$(mktemp)"
AUDIT_HIGH_OUTPUT_FILE="$(mktemp)"

# Keep checks deterministic and local.
"$SMOKE_BIN" --ping 127.0.0.1 --raw >"$PING_OUTPUT_FILE" 2>&1
"$SMOKE_BIN" --dns localhost --raw >"$DNS_OUTPUT_FILE" 2>&1
"$SMOKE_BIN" --network 127.0.0.1/32 --ports 1-32 --timeout 1 --threads 1 --audit-open-ports >"$AUDIT_OUTPUT_FILE" 2>&1
"$SMOKE_BIN" --network 127.0.0.1/32 --ports 1-32 --timeout 1 --threads 1 --audit-open-ports --audit-min-severity high >"$AUDIT_HIGH_OUTPUT_FILE" 2>&1

if ! rg -q "Ping: 127.0.0.1" "$PING_OUTPUT_FILE"; then
  echo "Smoke failed: expected ping summary header"
  exit 1
fi
if ! rg -q "raw ping output" "$PING_OUTPUT_FILE"; then
  echo "Smoke failed: expected raw ping section with --raw"
  exit 1
fi
if ! rg -q "DNS: localhost" "$DNS_OUTPUT_FILE"; then
  echo "Smoke failed: expected DNS summary header"
  exit 1
fi
if ! rg -q "raw dns output" "$DNS_OUTPUT_FILE"; then
  echo "Smoke failed: expected raw dns section with --raw"
  exit 1
fi
if [[ ! -s "$AUDIT_OUTPUT_FILE" ]]; then
  echo "Smoke failed: expected non-empty audit output"
  exit 1
fi
if [[ ! -s "$AUDIT_HIGH_OUTPUT_FILE" ]]; then
  echo "Smoke failed: expected non-empty audit output with high filter"
  exit 1
fi

echo "Smoke passed: CLI tools mode outputs expected sections."
