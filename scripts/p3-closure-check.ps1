param()

$ErrorActionPreference = "Stop"

function Assert-LastExitCode {
    param([string]$Step)
    if ($LASTEXITCODE -ne 0) {
        throw "$Step failed with exit code $LASTEXITCODE"
    }
}

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $root

Write-Host "== P3 closure check (Windows) ==" -ForegroundColor Cyan

Write-Host "[1/5] go test ./..." -ForegroundColor Yellow
go test ./...
Assert-LastExitCode "go test ./..."

Write-Host "[2/5] go test -tags=integration ./..." -ForegroundColor Yellow
go test -tags=integration ./...
Assert-LastExitCode "go test -tags=integration ./..."

Write-Host "[3/5] golden check" -ForegroundColor Yellow
go test ./internal/display -run Golden
Assert-LastExitCode "golden check"

Write-Host "[4/5] perf benchmark (FormatResultsAsTextLarge)" -ForegroundColor Yellow
go test ./internal/display -run ^$ -bench BenchmarkFormatResultsAsTextLarge -benchmem
Assert-LastExitCode "perf benchmark"

Write-Host "[5/5] p2 closure baseline (tools/smoke)" -ForegroundColor Yellow
.\scripts\p2-closure-check.ps1

Write-Host "P3 closure check passed." -ForegroundColor Green
