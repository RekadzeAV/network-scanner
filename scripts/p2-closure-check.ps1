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

Write-Host "== P2 closure check (Windows) ==" -ForegroundColor Cyan

Write-Host "[1/5] go test ./..." -ForegroundColor Yellow
go test ./...
Assert-LastExitCode "go test ./..."

Write-Host "[2/5] smoke-cli-no-topology.ps1" -ForegroundColor Yellow
.\scripts\smoke-cli-no-topology.ps1

Write-Host "[3/5] smoke-cli-topology.ps1" -ForegroundColor Yellow
.\scripts\smoke-cli-topology.ps1

Write-Host "[4/5] smoke-cli-tools.ps1" -ForegroundColor Yellow
.\scripts\smoke-cli-tools.ps1

Write-Host "[5/5] focused P2 flags sanity" -ForegroundColor Yellow
$smokeExe = Join-Path ([System.IO.Path]::GetTempPath()) ("network-scanner-smoke-" + [guid]::NewGuid().ToString() + ".exe")
try {
    go build -o $smokeExe .\cmd\network-scanner
    Assert-LastExitCode "go build smoke binary"

    & $smokeExe --network 127.0.0.1/32 --ports 80,443 --timeout 1 --grab-banners --show-raw-banners --os-detect-active *> $null
    Assert-LastExitCode "P2 flags scan command"

    $prevErrAction = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    & $smokeExe --wol-mac invalid-mac *> $null
    $ErrorActionPreference = $prevErrAction
    if ($LASTEXITCODE -eq 0) {
        throw "P2 closure failed: invalid WOL MAC should return non-zero exit code"
    }
}
finally {
    Remove-Item -Path $smokeExe -ErrorAction SilentlyContinue
}

Write-Host "P2 closure check passed." -ForegroundColor Green

