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

Write-Host "== Smoke: CLI with topology ==" -ForegroundColor Cyan

$smokeExe = Join-Path ([System.IO.Path]::GetTempPath()) ("network-scanner-smoke-" + [guid]::NewGuid().ToString() + ".exe")
go build -o $smokeExe .\cmd\network-scanner
Assert-LastExitCode "go build smoke binary"

$outputFile = [System.IO.Path]::GetTempFileName()
try {
    & $smokeExe --network 127.0.0.1/32 --timeout $TimeoutSec --ports 1-16 --topology *> $outputFile
    Assert-LastExitCode "topology smoke command"
    $output = Get-Content -Path $outputFile -Raw

    if ($output -notmatch "SNMP") {
        throw "Smoke failed: expected SNMP summary output in topology mode"
    }
}
finally {
    Remove-Item -Path $outputFile, $smokeExe -ErrorAction SilentlyContinue
}

Write-Host "Smoke passed: topology mode prints SNMP summary." -ForegroundColor Green
