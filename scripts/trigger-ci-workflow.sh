#!/usr/bin/env bash
set -euo pipefail

OWNER="${1:-RekadzeAV}"
REPO="${2:-network-scanner}"
WORKFLOW_FILE="${3:-ci.yml}"
REF="${4:-main}"
TIMEOUT_MINUTES="${5:-30}"
POLL_SECONDS="${6:-15}"

if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 is required for JSON parsing in this script." >&2
  exit 1
fi

if [[ -z "${GITHUB_TOKEN:-}" ]]; then
  echo "GITHUB_TOKEN is not set. Create a token with repo/workflow permissions and export it before running." >&2
  exit 1
fi

API_BASE="https://api.github.com/repos/${OWNER}/${REPO}"
DISPATCH_URL="${API_BASE}/actions/workflows/${WORKFLOW_FILE}/dispatches"
RUNS_URL="${API_BASE}/actions/workflows/${WORKFLOW_FILE}/runs?per_page=20"

echo "Dispatching workflow '${WORKFLOW_FILE}' on ref '${REF}'..."
curl -fsSL -X POST \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer ${GITHUB_TOKEN}" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  "${DISPATCH_URL}" \
  -d "{\"ref\":\"${REF}\"}" >/dev/null

echo "Waiting for the run to appear..."
end_ts=$(( $(date +%s) + TIMEOUT_MINUTES * 60 ))
run_id=""
run_url=""

while [[ "$(date +%s)" -lt "${end_ts}" ]]; do
  json="$(curl -fsSL \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: Bearer ${GITHUB_TOKEN}" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    "${RUNS_URL}")"

  run_id="$(printf '%s' "${json}" | python3 -c '
import json,sys
ref=sys.argv[1]
data=json.load(sys.stdin)
for run in data.get("workflow_runs", []):
    if run.get("head_branch") == ref:
        print(run.get("id"))
        break
' "${REF}")"

  if [[ -n "${run_id}" ]]; then
    run_url="https://github.com/${OWNER}/${REPO}/actions/runs/${run_id}"
    break
  fi

  sleep "${POLL_SECONDS}"
done

if [[ -z "${run_id}" ]]; then
  echo "Timed out while waiting for a new workflow run to appear." >&2
  exit 1
fi

echo "Run detected: id=${run_id}"
echo "URL: ${run_url}"
echo "Waiting for completion..."

while [[ "$(date +%s)" -lt "${end_ts}" ]]; do
  run_json="$(curl -fsSL \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: Bearer ${GITHUB_TOKEN}" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    "${API_BASE}/actions/runs/${run_id}")"

  status="$(printf '%s' "${run_json}" | python3 -c 'import json,sys; d=json.load(sys.stdin); print(d.get("status",""))')"
  conclusion="$(printf '%s' "${run_json}" | python3 -c 'import json,sys; d=json.load(sys.stdin); c=d.get("conclusion"); print("" if c is None else c)')"
  updated="$(printf '%s' "${run_json}" | python3 -c 'import json,sys; d=json.load(sys.stdin); print(d.get("updated_at",""))')"

  echo "status=${status} conclusion=${conclusion} updated=${updated}"
  if [[ "${status}" == "completed" ]]; then
    echo ""
    echo "Final run URL: ${run_url}"
    if [[ "${conclusion}" == "success" ]]; then
      jobs_json="$(curl -fsSL \
        -H "Accept: application/vnd.github+json" \
        -H "Authorization: Bearer ${GITHUB_TOKEN}" \
        -H "X-GitHub-Api-Version: 2022-11-28" \
        "${API_BASE}/actions/runs/${run_id}/jobs?per_page=100")"

      check_result="$(printf '%s' "${jobs_json}" | python3 -c '
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

      for token in ${check_result}; do
        case "${token}" in
          TEST_TOTAL=*) TEST_TOTAL="${token#*=}" ;;
          BUILD_TOTAL=*) BUILD_TOTAL="${token#*=}" ;;
          STAGE2_P1_TOTAL=*) STAGE2_P1_TOTAL="${token#*=}" ;;
          STAGE2_P3_TOTAL=*) STAGE2_P3_TOTAL="${token#*=}" ;;
          ALL=1) ALL_OK="YES" ;;
          ALL=0) ALL_OK="NO" ;;
        esac
      done

      echo "Required jobs check for P3 closure:"
      if [[ "${check_result}" == *"LINT=1"* ]]; then
        echo "- Lint: OK"
      else
        echo "- Lint: FAIL"
      fi
      if [[ "${check_result}" == *"TEST=1"* ]]; then
        echo "- Test matrix: OK (${TEST_TOTAL} jobs)"
      else
        echo "- Test matrix: FAIL (${TEST_TOTAL} jobs)"
      fi
      if [[ "${check_result}" == *"BUILD=1"* ]]; then
        echo "- Build and Smoke matrix: OK (${BUILD_TOTAL} jobs)"
      else
        echo "- Build and Smoke matrix: FAIL (${BUILD_TOTAL} jobs)"
      fi
      if [[ "${check_result}" == *"STAGE2_P1=1"* ]]; then
        echo "- Stage2 P1 Closure: OK (${STAGE2_P1_TOTAL} jobs)"
      else
        echo "- Stage2 P1 Closure: FAIL (${STAGE2_P1_TOTAL} jobs)"
      fi
      if [[ "${check_result}" == *"STAGE2_P3=1"* ]]; then
        echo "- Stage2 P3 Closure: OK (${STAGE2_P3_TOTAL} jobs)"
      else
        echo "- Stage2 P3 Closure: FAIL (${STAGE2_P3_TOTAL} jobs)"
      fi
      echo "- All required jobs green: ${ALL_OK}"

      if [[ "${check_result}" != *"ALL=1"* ]]; then
        echo "Workflow is success, but required jobs (Lint/Test/Build and Smoke/Stage2 closures) are not fully green." >&2
        exit 1
      fi

      echo "CI completed successfully; required jobs are green."
      exit 0
    fi
    echo "CI completed with conclusion '${conclusion}'." >&2
    exit 1
  fi
  sleep "${POLL_SECONDS}"
done

echo "Timed out waiting for workflow completion." >&2
exit 1
