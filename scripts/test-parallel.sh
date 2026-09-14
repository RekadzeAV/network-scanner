#!/bin/bash
set -euo pipefail

# Parallel test runner for Network Scanner
# Runs tests in parallel with timeout and colored output

cd "$(dirname "$0")/.."

echo "=============================================="
echo "  Network Scanner - Parallel Tests"
echo "=============================================="
echo ""

# Configuration
TEST_TIMEOUT=${TEST_TIMEOUT:-5m}
TEST_PARALLEL=${TEST_PARALLEL:-4}
VERBOSE=${VERBOSE:-0}

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test results
FAILED=0
PASSED=0
TOTAL=0

# Function to run tests for a package
run_test() {
    local pkg="$1"
    local desc="$2"
    local result_file="$3"
    
    if [ "$VERBOSE" = "1" ]; then
        echo -e "\n${YELLOW}Testing: ${desc}${NC}"
        echo "Package: ${pkg}"
    fi
    
    # Run tests with timeout
    if timeout "${TEST_TIMEOUT}" go test -parallel="${TEST_PARALLEL}" -timeout="${TEST_TIMEOUT}" "${pkg}" > /tmp/test_$(echo ${pkg} | tr '/' '_').log 2>&1; then
        echo "PASS" > "$result_file"
        if [ "$VERBOSE" = "0" ]; then
            echo -e "${GREEN}✓${NC} ${desc}"
        fi
        return 0
    else
        echo "FAIL" > "$result_file"
        echo -e "${RED}✗${NC} ${desc}"
        echo "  Log: /tmp/test_$(echo ${pkg} | tr '/' '_').log"
        if [ "$VERBOSE" = "1" ]; then
            cat /tmp/test_$(echo ${pkg} | tr '/' '_').log
        fi
        return 1
    fi
}

# Discover all test packages (excluding vendor, gui, and integration)
echo "==> Discovering test packages..."
PACKAGES=()
while IFS= read -r line; do
    PACKAGES+=("$line")
done < <(go list ./... | grep -v '/vendor/' | grep -v '/internal/gui' | grep -v '/integration' || true)

echo "Found ${#PACKAGES[@]} packages to test"
echo ""

# Run tests in parallel
echo "==> Running tests with ${TEST_PARALLEL} parallel workers..."
echo ""

PIDS=()
RESULTS_DIR=$(mktemp -d)

for pkg in "${PACKAGES[@]}"; do
    desc=$(echo "$pkg" | sed 's|network-scanner/||g')
    result_file="${RESULTS_DIR}/$(echo ${pkg} | tr '/' '_').result"
    run_test "$pkg" "$desc" "$result_file" &
    PIDS+=($!)
done

# Wait for all tests to complete
echo "==> Waiting for tests to complete..."
for pid in "${PIDS[@]}"; do
    wait "$pid" || true
done

# Count results
for result_file in "${RESULTS_DIR}"/*.result; do
    if [ -f "$result_file" ]; then
        TOTAL=$((TOTAL + 1))
        status=$(cat "$result_file")
        if [ "$status" = "PASS" ]; then
            PASSED=$((PASSED + 1))
        else
            FAILED=$((FAILED + 1))
        fi
    fi
done

# Cleanup
rm -rf "${RESULTS_DIR}"
rm -f /tmp/test_*.log

# Summary
echo ""
echo "=============================================="
echo "  Test Summary"
echo "=============================================="
echo -e "  Total:   ${TOTAL}"
echo -e "  ${GREEN}Passed:  ${PASSED}${NC}"
if [ "$FAILED" -gt 0 ]; then
    echo -e "  ${RED}Failed:  ${FAILED}${NC}"
else
    echo -e "  Failed:  ${FAILED}"
fi
echo "=============================================="

if [ "$FAILED" -gt 0 ]; then
    exit 1
fi

echo -e "\n${GREEN}All tests passed!${NC}"
