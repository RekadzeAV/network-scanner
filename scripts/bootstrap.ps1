# Bootstrap script for Network Scanner
# Checks environment, installs dependencies, builds and tests

Param(
    [switch]$SkipBuild,
    [switch]$SkipTest,
    [switch]$Verbose
)

$ErrorActionPreference = "Stop"

# Ensure GOPATH/bin is in PATH for installed tools
$env:PATH += ";$env:GOPATH\bin"

# Graphviz path for topology export
$graphvizPath = "C:\Program Files\Graphviz\bin"
if (Test-Path $graphvizPath) {
    $env:PATH += ";$graphvizPath"
}

Write-Host "=============================================="
Write-Host "  Network Scanner - Bootstrap"
Write-Host "=============================================="
Write-Host ""

# ============================================
# Environment Checks
# ============================================

function Check-Tool {
    param(
        [string]$Name,
        [string]$Command,
        [string]$VersionArg,
        [bool]$Required = $true,
        [string]$InstallGuide = ""
    )
    
    $tool = Get-Command $Command -ErrorAction SilentlyContinue
    if ($tool) {
        $version = & $Command $VersionArg 2>&1 | Select-Object -First 1
        Write-Host "[OK]   $Name - $version" -ForegroundColor Green
        return $true
    } else {
        if ($Required) {
            Write-Host "[FAIL] $Name is not installed or not in PATH" -ForegroundColor Red
            if ($InstallGuide) {
                Write-Host "       $InstallGuide" -ForegroundColor Yellow
            }
            throw "Required tool '$Name' is missing. Installation is required."
        } else {
            Write-Host "[WARN] $Name is not installed (optional)" -ForegroundColor Yellow
            if ($InstallGuide) {
                Write-Host "       $InstallGuide" -ForegroundColor Yellow
            }
            return $false
        }
    }
}

Write-Host "==> Checking required tools..."
Write-Host ""

# Check Go
$goVersion = go version 2>&1 | Select-Object -First 1
if ($goVersion -match "go(\d+\.\d+)") {
    $versionNum = $matches[1]
    Write-Host "[OK]   Go - $goVersion" -ForegroundColor Green
    # Compare versions (simple check for 1.24+)
    if ([version]$versionNum -lt [version]"1.24") {
        Write-Host "[WARN] Go version should be 1.24 or higher" -ForegroundColor Yellow
    }
} else {
    Write-Host "[FAIL] Go is not installed or not in PATH" -ForegroundColor Red
    throw "Go 1.24+ is required. Install from https://go.dev/dl/"
}

# Check GCC (required for GUI with CGO)
$gcc = Get-Command gcc -ErrorAction SilentlyContinue
if ($gcc) {
    $gccVersion = gcc --version 2>&1 | Select-Object -First 1
    Write-Host "[OK]   GCC - $gccVersion" -ForegroundColor Green
} else {
    Write-Host "[WARN] GCC not found (required for GUI build with CGO)" -ForegroundColor Yellow
    Write-Host "       Install: winget install BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source" -ForegroundColor Yellow
}

# Check CGO_ENABLED
$cgoEnabled = go env CGO_ENABLED
if ($cgoEnabled -eq "1") {
    Write-Host "[OK]   CGO_ENABLED - $cgoEnabled" -ForegroundColor Green
} else {
    Write-Host "[WARN] CGO_ENABLED=$cgoEnabled (GUI build may fail)" -ForegroundColor Yellow
    Write-Host "       Set: `$env:CGO_ENABLED = `"1`"" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "==> Checking optional development tools..."
Write-Host ""

# Check golangci-lint (optional but recommended)
Check-Tool -Name "golangci-lint" -Command "golangci-lint" -VersionArg "version" -Required $false -InstallGuide "Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

# Check govulncheck (optional but recommended)
if (Get-Command govulncheck -ErrorAction SilentlyContinue) {
    Write-Host "[OK]   govulncheck - installed" -ForegroundColor Green
} else {
    Write-Host "[WARN] govulncheck not installed (optional)" -ForegroundColor Yellow
    Write-Host "       Install: go install golang.org/x/vuln/cmd/govulncheck@latest" -ForegroundColor Yellow
}

# Check graphviz/dot (optional, needed for topology export)
try {
    $dotVersion = dot --version 2>&1 | Select-Object -First 1
    Write-Host "[OK]   Graphviz (dot) - $dotVersion" -ForegroundColor Green
} catch {
    Write-Host "[WARN] Graphviz (dot) not found (needed for topology PNG/SVG export)" -ForegroundColor Yellow
    Write-Host "       Install: winget install Graphviz.Graphviz" -ForegroundColor Yellow
}

Write-Host ""

# ============================================
# Download Dependencies
# ============================================

Write-Host "==> Downloading module dependencies..."
go mod download
if ($LASTEXITCODE -ne 0) {
    throw "Failed to download dependencies"
}
Write-Host "[OK]   Dependencies downloaded" -ForegroundColor Green

# Verify modules
Write-Host "==> Verifying module integrity..."
go mod verify
if ($LASTEXITCODE -ne 0) {
    throw "Module verification failed"
}
Write-Host "[OK]   Modules verified" -ForegroundColor Green

Write-Host ""

# ============================================
# Build
# ============================================

if (-not $SkipBuild) {
    Write-Host "==> Building CLI binary..."
    go build -o network-scanner.exe ./cmd/network-scanner
    if ($LASTEXITCODE -ne 0) {
        throw "CLI build failed"
    }
    Write-Host "[OK]   CLI binary: network-scanner.exe" -ForegroundColor Green

    Write-Host "==> Building GUI binary..."
    go build -o network-scanner-gui.exe ./cmd/gui
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[WARN] GUI build failed (may require CGO/GCC)" -ForegroundColor Yellow
    } else {
        Write-Host "[OK]   GUI binary: network-scanner-gui.exe" -ForegroundColor Green
    }

    Write-Host ""
}

# ============================================
# Tests
# ============================================

if (-not $SkipTest) {
    Write-Host "==> Running tests..."
    if ($Verbose) {
        go test -v ./...
    } else {
        go test ./...
    }
    if ($LASTEXITCODE -ne 0) {
        throw "Tests failed"
    }
    Write-Host "[OK]   All tests passed" -ForegroundColor Green
    Write-Host ""
}

# ============================================
# Summary
# ============================================

Write-Host "=============================================="
Write-Host "  Bootstrap completed successfully!"
Write-Host "=============================================="
Write-Host ""
Write-Host "Next steps:"
Write-Host "  - CLI: .\network-scanner.exe"
Write-Host "  - GUI: .\network-scanner-gui.exe (if built)"
Write-Host "  - Lint: golangci-lint run"
Write-Host "  - Security: govulncheck ./..."
Write-Host ""
