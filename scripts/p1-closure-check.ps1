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

Write-Host "== P1 closure check (Windows) ==" -ForegroundColor Cyan

Write-Host "[1/4] go test ./..." -ForegroundColor Yellow
go test ./...
Assert-LastExitCode "go test ./..."

Write-Host "[2/4] smoke-cli-no-topology.ps1" -ForegroundColor Yellow
.\scripts\smoke-cli-no-topology.ps1

Write-Host "[3/4] smoke-cli-topology.ps1" -ForegroundColor Yellow
.\scripts\smoke-cli-topology.ps1

Write-Host "[4/4] smoke-cli-tools.ps1" -ForegroundColor Yellow
.\scripts\smoke-cli-tools.ps1

Write-Host "P1 closure check passed." -ForegroundColor Green
