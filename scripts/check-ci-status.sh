#!/usr/bin/env bash
set -euo pipefail

OWNER="${1:-RekadzeAV}"
REPO="${2:-network-scanner}"
WORKFLOW_FILE="${3:-ci.yml}"
LIMIT="${4:-10}"

if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 is required for JSON parsing in this script." >&2
  exit 1
fi

URL="https://api.github.com/repos/${OWNER}/${REPO}/actions/workflows/${WORKFLOW_FILE}/runs?per_page=${LIMIT}"
JSON="$(curl -fsSL "$URL")"

echo "Recent CI runs (${OWNER}/${REPO}, workflow=${WORKFLOW_FILE}):"
printf '%s' "$JSON" | python3 -c '
import json,sys
data=json.load(sys.stdin)
for run in data.get("workflow_runs", []):
    print(f"- id={run.get(\"id\")} status={run.get(\"status\")} conclusion={run.get(\"conclusion\")} updated={run.get(\"updated_at\")}")
    print(f"  {run.get(\"html_url\")}")
'

SUCCESS_INFO="$(printf '%s' "$JSON" | python3 -c '
import json,sys
data=json.load(sys.stdin)
for run in data.get("workflow_runs", []):
    if run.get("status") == "completed" and run.get("conclusion") == "success":
        print(f"{run.get(\"id\")}|{run.get(\"html_url\")}")
        break
')"

if [[ -z "${SUCCESS_INFO}" ]]; then
  echo "No successful run found in recent history."
  exit 0
fi

SUCCESS_RUN_ID="${SUCCESS_INFO%%|*}"
SUCCESS_URL="${SUCCESS_INFO#*|}"

echo ""
echo "Latest successful run URL:"
echo "${SUCCESS_URL}"

JOBS_URL="https://api.github.com/repos/${OWNER}/${REPO}/actions/runs/${SUCCESS_RUN_ID}/jobs?per_page=100"
JOBS_JSON="$(curl -fsSL "${JOBS_URL}")"
CHECK_RESULT="$(printf '%s' "$JOBS_JSON" | python3 -c '
import json,sys
jobs=json.load(sys.stdin).get("jobs", [])
lint_ok=any(j.get("name")=="Lint" and j.get("conclusion")=="success" for j in jobs)
tests=[j for j in jobs if (j.get("name") or "").startswith("Test")]
builds=[j for j in jobs if (j.get("name") or "").startswith("Build and Smoke")]
stage2=[j for j in jobs if (j.get("name") or "").startswith("Stage2 P1 Closure")]
stage2_p3=[j for j in jobs if (j.get("name") or "").startswith("Stage2 P3 Closure")]
test_ok=len(tests) >= 3 and all(j.get("conclusion")=="success" for j in tests)
build_ok=len(builds) >= 3 and all(j.get("conclusion")=="success" for j in builds)
stage2_ok=len(stage2) >= 1 and all(j.get("conclusion")=="success" for j in stage2)
stage2_p3_ok=len(stage2_p3) >= 1 and all(j.get("conclusion")=="success" for j in stage2_p3)
print(f"LINT={int(lint_ok)} TEST={int(test_ok)} TEST_TOTAL={len(tests)} BUILD={int(build_ok)} BUILD_TOTAL={len(builds)} STAGE2_P1={int(stage2_ok)} STAGE2_P1_TOTAL={len(stage2)} STAGE2_P3={int(stage2_p3_ok)} STAGE2_P3_TOTAL={len(stage2_p3)} ALL={int(lint_ok and test_ok and build_ok and stage2_ok and stage2_p3_ok)}")
')"

echo ""
echo "Required jobs check for P3 closure:"
for token in ${CHECK_RESULT}; do
  case "${token}" in
    LINT=1) echo "- Lint: OK" ;;
    LINT=0) echo "- Lint: FAIL" ;;
    TEST=1) ;;
    TEST=0) ;;
    BUILD=1) ;;
    BUILD=0) ;;
    STAGE2_P1=1) ;;
    STAGE2_P1=0) ;;
    STAGE2_P3=1) ;;
    STAGE2_P3=0) ;;
    TEST_TOTAL=*) TEST_TOTAL="${token#*=}" ;;
    BUILD_TOTAL=*) BUILD_TOTAL="${token#*=}" ;;
    STAGE2_P1_TOTAL=*) STAGE2_P1_TOTAL="${token#*=}" ;;
    STAGE2_P3_TOTAL=*) STAGE2_P3_TOTAL="${token#*=}" ;;
    ALL=1) ALL_OK="YES" ;;
    ALL=0) ALL_OK="NO" ;;
  esac
done
if [[ "${CHECK_RESULT}" == *"TEST=1"* ]]; then
  echo "- Test matrix: OK (${TEST_TOTAL} jobs)"
else
  echo "- Test matrix: FAIL (${TEST_TOTAL} jobs)"
fi
if [[ "${CHECK_RESULT}" == *"BUILD=1"* ]]; then
  echo "- Build and Smoke matrix: OK (${BUILD_TOTAL} jobs)"
else
  echo "- Build and Smoke matrix: FAIL (${BUILD_TOTAL} jobs)"
fi
if [[ "${CHECK_RESULT}" == *"STAGE2_P1=1"* ]]; then
  echo "- Stage2 P1 Closure: OK (${STAGE2_P1_TOTAL} jobs)"
else
  echo "- Stage2 P1 Closure: FAIL (${STAGE2_P1_TOTAL} jobs)"
fi
if [[ "${CHECK_RESULT}" == *"STAGE2_P3=1"* ]]; then
  echo "- Stage2 P3 Closure: OK (${STAGE2_P3_TOTAL} jobs)"
else
  echo "- Stage2 P3 Closure: FAIL (${STAGE2_P3_TOTAL} jobs)"
fi
echo "- All required jobs green: ${ALL_OK}"
if [[ "${CHECK_RESULT}" != *"ALL=1"* ]]; then
  exit 1
fi
