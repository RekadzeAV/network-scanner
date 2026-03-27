param(
    [int]$TimeoutSec = 1
)

$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $root

Write-Host "== Smoke: CLI without topology ==" -ForegroundColor Cyan

go build -o ".\release\network-scanner-smoke.exe" .\cmd\network-scanner

$outputFile = [System.IO.Path]::GetTempFileName()
try {
    & ".\release\network-scanner-smoke.exe" --network 127.0.0.1/32 --timeout $TimeoutSec --ports 1-16 *> $outputFile
    $output = Get-Content -Path $outputFile -Raw

    if ($output -match "SNMP отчет") {
        throw "Smoke failed: SNMP report appears without --topology"
    }
}
finally {
    Remove-Item -Path $outputFile -ErrorAction SilentlyContinue
}

Write-Host "Smoke passed: baseline CLI path works without topology." -ForegroundColor Green
