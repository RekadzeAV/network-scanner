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

Write-Host "== Smoke: CLI tools mode (ping/dns) ==" -ForegroundColor Cyan

go test ./cmd/network-scanner -run WhoisUsesRDAPFallback -count=1
Assert-LastExitCode "whois cli e2e test"

$smokeExe = Join-Path ([System.IO.Path]::GetTempPath()) ("network-scanner-smoke-" + [guid]::NewGuid().ToString() + ".exe")
go build -o $smokeExe .\cmd\network-scanner
Assert-LastExitCode "go build smoke binary"

$pingOutputFile = [System.IO.Path]::GetTempFileName()
$dnsOutputFile = [System.IO.Path]::GetTempFileName()
$auditOutputFile = [System.IO.Path]::GetTempFileName()
$auditHighOutputFile = [System.IO.Path]::GetTempFileName()
try {
    & $smokeExe --ping 127.0.0.1 --raw *> $pingOutputFile
    Assert-LastExitCode "ping smoke command"
    & $smokeExe --dns localhost --raw *> $dnsOutputFile
    Assert-LastExitCode "dns smoke command"
    & $smokeExe --network 127.0.0.1/32 --ports 1-32 --timeout 1 --threads 1 --audit-open-ports *> $auditOutputFile
    Assert-LastExitCode "audit smoke command"
    & $smokeExe --network 127.0.0.1/32 --ports 1-32 --timeout 1 --threads 1 --audit-open-ports --audit-min-severity high *> $auditHighOutputFile
    Assert-LastExitCode "audit high smoke command"

    $pingOutput = Get-Content -Path $pingOutputFile -Raw
    $dnsOutput = Get-Content -Path $dnsOutputFile -Raw
    $auditOutput = Get-Content -Path $auditOutputFile -Raw
    $auditHighOutput = Get-Content -Path $auditHighOutputFile -Raw

    if ($pingOutput -notmatch "Ping: 127.0.0.1") {
        throw "Smoke failed: expected ping summary header"
    }
    if ($pingOutput -notmatch "raw ping output") {
        throw "Smoke failed: expected raw ping section with --raw"
    }
    if ($dnsOutput -notmatch "DNS: localhost") {
        throw "Smoke failed: expected DNS summary header"
    }
    if ($dnsOutput -notmatch "raw dns output") {
        throw "Smoke failed: expected raw dns section with --raw"
    }
    if ([string]::IsNullOrWhiteSpace($auditOutput)) {
        throw "Smoke failed: expected non-empty audit output"
    }
    if ([string]::IsNullOrWhiteSpace($auditHighOutput)) {
        throw "Smoke failed: expected non-empty audit output with high filter"
    }
}
finally {
    Remove-Item -Path $pingOutputFile, $dnsOutputFile, $auditOutputFile, $auditHighOutputFile, $smokeExe -ErrorAction SilentlyContinue
}

Write-Host "Smoke passed: CLI tools mode outputs expected sections." -ForegroundColor Green
