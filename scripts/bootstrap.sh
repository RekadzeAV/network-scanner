#!/usr/bin/env bash
set -euo pipefail

# Bootstrap script for Network Scanner
# Checks environment, installs dependencies, builds and tests

SKIP_BUILD=${SKIP_BUILD:-0}
SKIP_TEST=${SKIP_TEST:-0}
VERBOSE=${VERBOSE:-0}

echo "=============================================="
echo "  Network Scanner - Bootstrap"
echo "=============================================="
echo ""

# ============================================
# Environment Checks
# ============================================

check_tool() {
    local name="$1"
    local cmd="$2"
    local version_arg="$3"
    local required="${4:-true}"
    local install_guide="${5:-}"
    
    if command -v "$cmd" >/dev/null 2>&1; then
        local version
        version=$("$cmd" $version_arg 2>&1 | head -n1)
        echo "[OK]   $name - $version"
        return 0
    else
        if [ "$required" = "true" ]; then
            echo "[FAIL] $name is not installed or not in PATH"
            if [ -n "$install_guide" ]; then
                echo "       $install_guide"
            fi
            return 1
        else
            echo "[WARN] $name is not installed (optional)"
            if [ -n "$install_guide" ]; then
                echo "       $install_guide"
            fi
            return 0
        fi
    fi
}

echo "==> Checking required tools..."
echo ""

# Check Go
if command -v go >/dev/null 2>&1; then
    go_version=$(go version | head -n1)
    echo "[OK]   Go - $go_version"
else
    echo "[FAIL] Go is not installed or not in PATH"
    echo "       Install from https://go.dev/dl/"
    exit 1
fi

# Check GCC (required for GUI with CGO)
if command -v gcc >/dev/null 2>&1; then
    gcc_version=$(gcc --version | head -n1)
    echo "[OK]   GCC - $gcc_version"
else
    echo "[WARN] GCC not found (required for GUI build with CGO)"
    echo "       Install: sudo apt install gcc (Debian/Ubuntu) or sudo yum install gcc (RHEL/CentOS)"
fi

# Check CGO_ENABLED
cgo_enabled=$(go env CGO_ENABLED)
if [ "$cgo_enabled" = "1" ]; then
    echo "[OK]   CGO_ENABLED - $cgo_enabled"
else
    echo "[WARN] CGO_ENABLED=$cgo_enabled (GUI build may fail)"
    echo "       Set: export CGO_ENABLED=1"
fi

echo ""
echo "==> Checking optional development tools..."
echo ""

# Check golangci-lint
check_tool "golangci-lint" "golangci-lint" "version" "false" "Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" || true

# Check govulncheck
if command -v govulncheck >/dev/null 2>&1; then
    echo "[OK]   govulncheck - installed"
else
    echo "[WARN] govulncheck not installed (optional)"
    echo "       Install: go install golang.org/x/vuln/cmd/govulncheck@latest"
fi

# Check graphviz/dot
if command -v dot >/dev/null 2>&1; then
    dot_version=$(dot --version 2>&1 | head -n1)
    echo "[OK]   Graphviz (dot) - $dot_version"
else
    echo "[WARN] Graphviz (dot) not found (needed for topology PNG/SVG export)"
    echo "       Install: sudo apt install graphviz (Debian/Ubuntu) or sudo yum install graphviz (RHEL/CentOS)"
fi

echo ""

# ============================================
# Download Dependencies
# ============================================

echo "==> Downloading module dependencies..."
go mod download || { echo "Failed to download dependencies"; exit 1; }
echo "[OK]   Dependencies downloaded"

# Verify modules
echo "==> Verifying module integrity..."
go mod verify || { echo "Module verification failed"; exit 1; }
echo "[OK]   Modules verified"

echo ""

# ============================================
# Build
# ============================================

if [ "$SKIP_BUILD" = "0" ]; then
    echo "==> Building CLI binary..."
    go build -o network-scanner ./cmd/network-scanner || { echo "CLI build failed"; exit 1; }
    echo "[OK]   CLI binary: network-scanner"

    echo "==> Building GUI binary..."
    if go build -o network-scanner-gui ./cmd/gui; then
        echo "[OK]   GUI binary: network-scanner-gui"
    else
        echo "[WARN] GUI build failed (may require CGO/GCC)"
    fi

    echo ""
fi

# ============================================
# Tests
# ============================================

if [ "$SKIP_TEST" = "0" ]; then
    echo "==> Running tests..."
    if [ "$VERBOSE" = "1" ]; then
        go test -v ./... || { echo "Tests failed"; exit 1; }
    else
        go test ./... || { echo "Tests failed"; exit 1; }
    fi
    echo "[OK]   All tests passed"
    echo ""
fi

# ============================================
# Summary
# ============================================

echo "=============================================="
echo "  Bootstrap completed successfully!"
echo "=============================================="
echo ""
echo "Next steps:"
echo "  - CLI: ./network-scanner"
echo "  - GUI: ./network-scanner-gui (if built)"
echo "  - Lint: golangci-lint run"
echo "  - Security: govulncheck ./..."
echo ""
