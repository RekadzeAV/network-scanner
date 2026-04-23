param(
    [int]$TimeoutSec = 1
)

$ErrorActionPreference = "Stop"

function Assert-LastExitCode {
    param([string]$Step)
    if ($LASTEXITCODE -ne 0) {
        throw "$Step failed with exit code $LASTEXITCODE"
    }
}

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $root

Write-Host "== Smoke: CLI without topology ==" -ForegroundColor Cyan

$smokeExe = Join-Path ([System.IO.Path]::GetTempPath()) ("network-scanner-smoke-" + [guid]::NewGuid().ToString() + ".exe")
go build -o $smokeExe .\cmd\network-scanner
Assert-LastExitCode "go build smoke binary"

$outputFile = [System.IO.Path]::GetTempFileName()
try {
    & $smokeExe --network 127.0.0.1/32 --timeout $TimeoutSec --ports 1-16 --os-detect-active *> $outputFile
    Assert-LastExitCode "no-topology smoke command"
    $output = Get-Content -Path $outputFile -Raw

    if ($output -match "SNMP отчет") {
        throw "Smoke failed: SNMP report appears without --topology"
    }
}
finally {
    Remove-Item -Path $outputFile, $smokeExe -ErrorAction SilentlyContinue
}

Write-Host "Smoke passed: baseline CLI path works without topology." -ForegroundColor Green
