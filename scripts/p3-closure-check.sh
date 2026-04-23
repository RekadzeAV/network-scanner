#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

echo "== P3 closure check (Unix) =="

echo "[1/5] go test ./..."
go test ./...

echo "[2/5] go test -tags=integration ./..."
go test -tags=integration ./...

echo "[3/5] golden check"
go test ./internal/display -run Golden

echo "[4/5] perf benchmark (FormatResultsAsTextLarge)"
go test ./internal/display -run ^$ -bench BenchmarkFormatResultsAsTextLarge -benchmem

echo "[5/5] p2 closure baseline (tools/smoke)"
./scripts/p2-closure-check.sh

echo "P3 closure check passed."
