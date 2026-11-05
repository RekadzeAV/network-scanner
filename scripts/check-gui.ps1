param(
    [switch]$SkipTests,
    [switch]$NoColor
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# --- Цвета -------------------------------------------------------------------

function Write-OK  { param([string]$Text); Write-Host "[OK]  $Text"  -ForegroundColor Green }
function Write-FAIL { param([string]$Text); Write-Host "[FAIL] $Text" -ForegroundColor Red }
function Write-INFO { param([string]$Text); Write-Host "[--]  $Text" -ForegroundColor Cyan }
function Write-SEP  { param([string]$Text); Write-Host "=== $Text ===" -ForegroundColor Magenta }

# --- Go installed? -----------------------------------------------------------

$goCmd = Get-Command go -ErrorAction SilentlyContinue
if (-not $goCmd) { throw 'go not found in PATH' }

# --- Root --------------------------------------------------------------------

$Root = $PWD.Path
$GuiDir = Join-Path $Root 'internal\gui'

# --- Counters ----------------------------------------------------------------

$TotalChecks  = 0
$PassedChecks = 0

function Report-OK  { param([string]$Name); Write-OK $Name; $script:TotalChecks++; $script:PassedChecks++ }
function Report-FAIL { param([string]$Name); Write-FAIL $Name; $script:TotalChecks++ }

# ============================================================================
# (a) color.RGBA{ — only theme_tokens.go
# ============================================================================
Write-SEP 'Point (a): color.RGBA literals'

$failCountA = 0
$goFiles = Get-ChildItem -Path $GuiDir -Filter '*.go' -Recurse
foreach ($f in $goFiles) {
    $name = $f.Name
    if ($name -match '_test\.go$') { continue }
    if ($name -eq 'theme_tokens.go') { continue }

    $lines = Get-Content $f.FullName
    $ln = 0
    foreach ($line in $lines) {
        $ln++
        if ($line -match 'color\.RGBA\{') {
            $failCountA++
            $rel = $f.FullName.Substring($Root.Length).TrimStart('\')
            Write-FAIL "  $rel : $ln"
        }
    }
}

if ($failCountA -eq 0) {
    Report-OK 'color.RGBA{} only in theme_tokens.go'
}
else {
    Report-FAIL "color.RGBA{} only in theme_tokens.go ($failCountA violation(s) found)"
}

# ============================================================================
# (b) gofmt -l internal/gui
# ============================================================================
Write-SEP 'Point (b): gofmt'

Set-Location $Root
$gofmtOut = gofmt -l internal\gui 2>&1

$gofmtClean = $true
if ($null -ne $gofmtOut -and "$gofmtOut".Trim() -ne '') {
    $gofmtClean = $false
}

if ($gofmtClean) {
    Report-OK 'gofmt -l internal\gui -- clean'
}
else {
    Report-FAIL "gofmt -l internal\gui -- not formatted: $gofmtOut"
}

# ============================================================================
# (c) go vet
# ============================================================================
Write-SEP 'Point (c): go vet'

$vetOut = go vet ./internal/gui/... 2>&1
$vetOk  = ($LASTEXITCODE -eq 0)

if ($vetOk) {
    Report-OK 'go vet ./internal/gui/...'
}
else {
    Report-FAIL "go vet ./internal/gui/...: $vetOut"
}

# ============================================================================
# (d) go test ./internal/gui/... -short
# ============================================================================
Write-SEP 'Point (d): go test ./internal/gui/...'

if ($SkipTests) {
    Write-INFO 'Point (d) skipped (-SkipTests)'
}
else {
    $testOut = go test ./internal/gui/... -count=1 -short 2>&1
    $testOk  = ($LASTEXITCODE -eq 0)

    if ($testOk) {
        Report-OK 'go test ./internal/gui/... -count=1 -short'
    }
    else {
        Report-FAIL "go test ./internal/gui/... -count=1 -short: $testOut"
    }
}

# ============================================================================
# (e) TestWCAGSweep
# ============================================================================
Write-SEP 'Point (e): go test TestWCAGSweep'

if ($SkipTests) {
    Write-INFO 'Point (e) skipped (-SkipTests)'
}
else {
    $wcagOut = go test ./internal/gui/ -run TestWCAGSweep -count=1 -short 2>&1
    $wcagOk  = ($LASTEXITCODE -eq 0)

    if ($wcagOk) {
        Report-OK 'go test -run TestWCAGSweep -count=1 -short'
    }
    else {
        Report-FAIL "go test -run TestWCAGSweep -count=1 -short: $wcagOut"
    }
}

# ============================================================================
# Summary
# ============================================================================
Write-SEP 'SUMMARY'

$failed = $TotalChecks - $PassedChecks

if ($failed -eq 0) {
    Write-OK "All $TotalChecks checks passed"
    exit 0
}
else {
    Write-FAIL "$failed of $TotalChecks checks FAILED"
    exit 1
}
