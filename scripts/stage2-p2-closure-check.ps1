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

Write-Host "== Stage2 P2 closure check (Windows) ==" -ForegroundColor Cyan

Write-Host "[1/5] go test ./internal/risksignature ./internal/devicecontrol ./internal/report ./internal/gui" -ForegroundColor Yellow
go test ./internal/risksignature ./internal/devicecontrol ./internal/report ./internal/gui
Assert-LastExitCode "go test internal packages"

Write-Host "[2/5] smoke-cli-tools.ps1" -ForegroundColor Yellow
.\scripts\smoke-cli-tools.ps1

Write-Host "[3/5] build smoke binary" -ForegroundColor Yellow
$smokeExe = Join-Path ([System.IO.Path]::GetTempPath()) ("network-scanner-stage2-p2-" + [guid]::NewGuid().ToString() + ".exe")
$reportFile = Join-Path ([System.IO.Path]::GetTempPath()) ("security-report-stage2-p2-" + [guid]::NewGuid().ToString() + ".html")
try {
    go build -o $smokeExe .\cmd\network-scanner
    Assert-LastExitCode "go build smoke binary"

    Write-Host "[4/5] risk-signatures + security-report sanity" -ForegroundColor Yellow
    & $smokeExe --network 127.0.0.1/32 --ports 22,80 --timeout 1 --threads 1 --risk-signatures --security-report-file $reportFile *> $null
    Assert-LastExitCode "risk-signatures + security-report sanity"

    $reportContent = Get-Content -Path $reportFile -Raw
    if ($reportContent -notmatch "CVE Findings") {
        throw "Security report does not contain 'CVE Findings'"
    }
    if ($reportContent -notmatch "Risk Signature Findings") {
        throw "Security report does not contain 'Risk Signature Findings'"
    }

    Write-Host "[5/5] device-control negative cases" -ForegroundColor Yellow
    $proc = Start-Process -FilePath $smokeExe -ArgumentList @(
        "--device-action", "reboot",
        "--device-target", "http://127.0.0.1",
        "--device-vendor", "generic-http"
    ) -NoNewWindow -Wait -PassThru
    if ($proc.ExitCode -eq 0) {
        throw "Expected reboot without --device-confirm to fail"
    }

    $proc = Start-Process -FilePath $smokeExe -ArgumentList @(
        "--device-action", "status"
    ) -NoNewWindow -Wait -PassThru
    if ($proc.ExitCode -eq 0) {
        throw "Expected status without --device-target to fail"
    }
}
finally {
    Remove-Item -Path $smokeExe -ErrorAction SilentlyContinue
    Remove-Item -Path $reportFile -ErrorAction SilentlyContinue
}

Write-Host "Stage2 P2 closure check passed." -ForegroundColor Green
