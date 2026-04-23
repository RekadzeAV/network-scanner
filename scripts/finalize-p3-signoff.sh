#!/usr/bin/env bash
set -euo pipefail

OWNER="${1:-RekadzeAV}"
REPO="${2:-network-scanner}"
WORKFLOW_FILE="${3:-ci.yml}"
CHECKLIST_PATH="${4:-docs/P3_CLOSURE_CHECKLIST.md}"
CONFIRMED_BY="${5:-TBD}"
DATE_VALUE="${6:-$(date +%F)}"

if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 is required for this script." >&2
  exit 1
fi

RUNS_URL="https://api.github.com/repos/${OWNER}/${REPO}/actions/workflows/${WORKFLOW_FILE}/runs?per_page=20"
RUNS_JSON="$(curl -fsSL "${RUNS_URL}")"

RUN_INFO="$(printf '%s' "${RUNS_JSON}" | python3 -c '
import json,sys
data=json.load(sys.stdin)
for run in data.get("workflow_runs", []):
    if run.get("status") == "completed" and run.get("conclusion") == "success":
        print(f"{run.get(\"id\")}|{run.get(\"html_url\")}")
        break
')"

if [[ -z "${RUN_INFO}" ]]; then
  echo "No successful CI run found for workflow '${WORKFLOW_FILE}'." >&2
  exit 1
fi

RUN_ID="${RUN_INFO%%|*}"
RUN_URL="${RUN_INFO#*|}"

JOBS_URL="https://api.github.com/repos/${OWNER}/${REPO}/actions/runs/${RUN_ID}/jobs?per_page=100"
JOBS_JSON="$(curl -fsSL "${JOBS_URL}")"
REQUIRED_OK="$(printf '%s' "${JOBS_JSON}" | python3 -c '
import json,sys
jobs=json.load(sys.stdin).get("jobs", [])
lint_ok=any(j.get("name")=="Lint" and j.get("conclusion")=="success" for j in jobs)
tests=[j for j in jobs if (j.get("name") or "").startswith("Test")]
builds=[j for j in jobs if (j.get("name") or "").startswith("Build and Smoke")]
test_ok=len(tests) >= 3 and all(j.get("conclusion")=="success" for j in tests)
build_ok=len(builds) >= 3 and all(j.get("conclusion")=="success" for j in builds)
print("1" if (lint_ok and test_ok and build_ok) else "0")
')"

if [[ "${REQUIRED_OK}" != "1" ]]; then
  echo "Latest successful run does not satisfy required jobs: Lint/Test*/Build and Smoke*." >&2
  exit 1
fi

python3 - "$CHECKLIST_PATH" "$DATE_VALUE" "$CONFIRMED_BY" "$RUN_URL" "$RUN_ID" <<'PY'
import sys
from pathlib import Path

checklist = Path(sys.argv[1])
date_value = sys.argv[2]
confirmed_by = sys.argv[3]
run_url = sys.argv[4]
run_id = sys.argv[5]

lines = checklist.read_text(encoding="utf-8").splitlines()

def set_first_backtick_value(line: str, value: str) -> str:
    first = line.find("`")
    if first < 0:
        return line
    second = line.find("`", first + 1)
    if second < 0:
        return line
    return line[:first + 1] + value + line[second:]

def replace_after_first_backtick_pair(line: str, tail: str) -> str:
    first = line.find("`")
    if first < 0:
        return line
    second = line.find("`", first + 1)
    if second < 0:
        return line
    return line[:second + 1] + tail

start = -1
for i, ln in enumerate(lines):
    if ln.strip() == "## P3 Final Sign-off":
        start = i
        break
if start < 0:
    raise SystemExit("P3 Final Sign-off section not found in checklist.")

end = len(lines) - 1
for i in range(start + 1, len(lines)):
    if lines[i].startswith("## "):
        end = i - 1
        break

bullet_idx = [i for i in range(start, end + 1) if lines[i].lstrip().startswith("- ")]
if len(bullet_idx) < 8:
    raise SystemExit("Unexpected P3 Final Sign-off format. Not enough bullet items.")

lines[bullet_idx[0]] = set_first_backtick_value(lines[bullet_idx[0]], "closed")
lines[bullet_idx[1]] = set_first_backtick_value(lines[bullet_idx[1]], date_value)
lines[bullet_idx[2]] = set_first_backtick_value(lines[bullet_idx[2]], confirmed_by)
lines[bullet_idx[3]] = set_first_backtick_value(lines[bullet_idx[3]], run_url)
lines[bullet_idx[6]] = set_first_backtick_value(lines[bullet_idx[6]], "closed")
lines[bullet_idx[6]] = replace_after_first_backtick_pair(
    lines[bullet_idx[6]],
    " - required jobs (`Lint`, `Test*`, `Build and Smoke*`) confirmed green.",
)
lines[bullet_idx[7]] = set_first_backtick_value(lines[bullet_idx[7]], "./scripts/finalize-p3-signoff.sh")
lines[bullet_idx[7]] = replace_after_first_backtick_pair(
    lines[bullet_idx[7]],
    f" - run `{run_id}` validated.",
)

checklist.write_text("\n".join(lines) + "\n", encoding="utf-8")
PY

echo "P3 checklist updated from successful CI run."
echo "Run URL: ${RUN_URL}"
