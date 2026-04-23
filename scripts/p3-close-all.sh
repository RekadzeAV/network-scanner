#!/usr/bin/env bash
set -euo pipefail

OWNER="${1:-RekadzeAV}"
REPO="${2:-network-scanner}"
WORKFLOW_FILE="${3:-ci.yml}"
REF="${4:-main}"
CONFIRMED_BY="${5:-TBD}"
TIMEOUT_MINUTES="${6:-30}"
POLL_SECONDS="${7:-15}"

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

echo "== P3 close all (Unix) =="

if [[ -z "${GITHUB_TOKEN:-}" ]]; then
  echo "GITHUB_TOKEN is not set. Export a token with workflow/repo access before running p3-close-all." >&2
  exit 1
fi
if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 is required for p3-close-all scripts." >&2
  exit 1
fi

echo "[1/3] Trigger CI workflow and wait for completion"
./scripts/trigger-ci-workflow.sh "$OWNER" "$REPO" "$WORKFLOW_FILE" "$REF" "$TIMEOUT_MINUTES" "$POLL_SECONDS"

echo "[2/3] Check latest successful CI status"
./scripts/check-ci-status.sh "$OWNER" "$REPO" "$WORKFLOW_FILE"

echo "[3/3] Finalize P3 sign-off in checklist"
./scripts/finalize-p3-signoff.sh "$OWNER" "$REPO" "$WORKFLOW_FILE" "docs/P3_CLOSURE_CHECKLIST.md" "$CONFIRMED_BY"

echo "P3 close-all flow completed successfully."
