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

Write-Host "== Stage2 P3 closure check (Windows) ==" -ForegroundColor Cyan

Write-Host "[1/5] go test ./internal/cve ./internal/report ./internal/remoteexec ./cmd/network-scanner" -ForegroundColor Yellow
go test ./internal/cve ./internal/report ./internal/remoteexec ./cmd/network-scanner
Assert-LastExitCode "go test stage2 p3 packages"

Write-Host "[2/5] build smoke binary" -ForegroundColor Yellow
$smokeExe = Join-Path ([System.IO.Path]::GetTempPath()) ("network-scanner-stage2-p3-" + [guid]::NewGuid().ToString() + ".exe")
$outFile = Join-Path ([System.IO.Path]::GetTempPath()) ("network-scanner-stage2-p3-out-" + [guid]::NewGuid().ToString() + ".txt")
$auditFile = Join-Path ([System.IO.Path]::GetTempPath()) ("network-scanner-stage2-p3-audit-" + [guid]::NewGuid().ToString() + ".jsonl")

try {
    go build -o $smokeExe .\cmd\network-scanner
    Assert-LastExitCode "go build smoke binary"

    Write-Host "[3/5] security-report auto(redacted) sanity" -ForegroundColor Yellow
    & $smokeExe --network 127.0.0.1/32 --ports 22,80 --timeout 1 --threads 1 --risk-signatures --security-report-file auto --security-report-redact=true *> $outFile
    Assert-LastExitCode "security report auto redacted"
    $outText = Get-Content $outFile -Raw
    if ($outText -notmatch 'security-report-redacted-.*\.html') {
        throw "Expected redacted auto report filename in output"
    }

    Write-Host "[4/5] security-report unredacted consent sanity" -ForegroundColor Yellow
    & $smokeExe --network 127.0.0.1/32 --ports 22,80 --timeout 1 --threads 1 --risk-signatures --security-report-file auto --security-report-redact=false --security-report-unsafe-consent I_UNDERSTAND_UNREDACTED_REPORT *> $outFile
    Assert-LastExitCode "security report auto unredacted"
    $outText = Get-Content $outFile -Raw
    if ($outText -notmatch 'security-report-unredacted-.*\.html') {
        throw "Expected unredacted auto report filename in output"
    }
    if ($outText -notmatch 'report-id=') {
        throw "Expected report-id marker in output"
    }

    Write-Host "[5/5] remote-exec dry-run strict policy sanity" -ForegroundColor Yellow
    & $smokeExe --remote-exec-transport ssh --remote-exec-target 127.0.0.1 --remote-exec-command hostname --remote-exec-policy-file config/remote-exec-policy.example.json --remote-exec-policy-strict --remote-exec-consent I_UNDERSTAND --remote-exec-dry-run=true --remote-exec-audit-log $auditFile *> $outFile
    $outText = Get-Content $outFile -Raw
    if ($outText -notmatch 'Remote exec policy ошибка|target is not in allowlist') {
        throw "Expected policy/allowlist guardrail output"
    }
}
finally {
    Remove-Item -Path $smokeExe -ErrorAction SilentlyContinue
    Remove-Item -Path $outFile -ErrorAction SilentlyContinue
    Remove-Item -Path $auditFile -ErrorAction SilentlyContinue
}

Write-Host "Stage2 P3 closure check passed." -ForegroundColor Green
