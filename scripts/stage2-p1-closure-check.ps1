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

Write-Host "== Stage2 P1 closure check (Windows) ==" -ForegroundColor Cyan

Write-Host "[1/5] go test ./internal/nettools ./internal/audit ./internal/gui" -ForegroundColor Yellow
go test ./internal/nettools ./internal/audit ./internal/gui
Assert-LastExitCode "go test internal packages"

Write-Host "[2/5] go test ./cmd/network-scanner -run Whois" -ForegroundColor Yellow
go test ./cmd/network-scanner -run Whois
Assert-LastExitCode "go test cmd/network-scanner whois"

Write-Host "[3/5] smoke-cli-tools.ps1" -ForegroundColor Yellow
.\scripts\smoke-cli-tools.ps1

Write-Host "[4/5] build smoke binary" -ForegroundColor Yellow
$smokeExe = Join-Path ([System.IO.Path]::GetTempPath()) ("network-scanner-stage2-p1-" + [guid]::NewGuid().ToString() + ".exe")
try {
    go build -o $smokeExe .\cmd\network-scanner
    Assert-LastExitCode "go build smoke binary"

    Write-Host "[5/5] audit-min-severity sanity" -ForegroundColor Yellow
    & $smokeExe --network 127.0.0.1/32 --ports 1-32 --timeout 1 --threads 1 --audit-open-ports --audit-min-severity critical *> $null
    Assert-LastExitCode "audit-min-severity critical"
    & $smokeExe --network 127.0.0.1/32 --ports 1-32 --timeout 1 --threads 1 --audit-open-ports --audit-min-severity high *> $null
    Assert-LastExitCode "audit-min-severity high"
}
finally {
    Remove-Item -Path $smokeExe -ErrorAction SilentlyContinue
}

Write-Host "Stage2 P1 closure check passed." -ForegroundColor Green
