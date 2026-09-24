# Docs generation: godoc HTML для внутренних пакетов (M3).
#
# Генерирует статический HTML-сайт godoc в docs/dev/godoc/.
# Требует: Go 1.25+ и `go install golang.org/x/tools/cmd/godoc@latest`
# (или использует `go doc` для терминального вывода, если godoc недоступен).
#
# Usage (PowerShell):
#   ./scripts/gen-godoc.ps1

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$outDir = Join-Path $repoRoot "docs/dev/godoc"

Write-Host "==> Generating godoc into $outDir"

if (-not (Test-Path $outDir)) {
    New-Item -ItemType Directory -Path $outDir -Force | Out-Null
}

# Пытаемся использовать классический godoc, если он установлен.
$godoc = Get-Command godoc -ErrorAction SilentlyContinue
if ($godoc) {
    Push-Location $repoRoot
    try {
        godoc -http=:0 -goroot=. -path="$repoRoot" -write_index_to=$outDir 2>&1 | Out-Null
        Write-Host "==> godoc index written to $outDir"
    } finally {
        Pop-Location
    }
    exit 0
}

Write-Warning "godoc not installed; falling back to 'go doc' text dump."
$packages = Get-ChildItem -Path (Join-Path $repoRoot "internal") -Directory |
    Where-Object { Test-Path (Join-Path $_.FullName "*.go") }

$dump = Join-Path $outDir "GODOC_DUMP.txt"
"# godoc dump generated $(Get-Date -Format o)" | Set-Content -Path $dump -Encoding UTF8

Push-Location $repoRoot
try {
    foreach ($pkg in $packages) {
        $importPath = "network-scanner/internal/$($pkg.Name)"
        Add-Content -Path $dump -Value "`n===== $importPath =====`n"
        go doc -all $importPath 2>&1 | Add-Content -Path $dump
    }
} finally {
    Pop-Location
}

Write-Host "==> Fallback dump written to $dump"
Write-Host "    Install godoc for full HTML: go install golang.org/x/tools/cmd/godoc@latest"
