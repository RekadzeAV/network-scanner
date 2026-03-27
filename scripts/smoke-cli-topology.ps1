param(
    [int]$TimeoutSec = 1
)

$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $root

Write-Host "== Smoke: CLI with topology ==" -ForegroundColor Cyan

go build -o ".\release\network-scanner-smoke.exe" .\cmd\network-scanner

$outputFile = [System.IO.Path]::GetTempFileName()
try {
    & ".\release\network-scanner-smoke.exe" --network 127.0.0.1/32 --timeout $TimeoutSec --ports 1-16 --topology *> $outputFile
    $output = Get-Content -Path $outputFile -Raw

    if ($output -notmatch "SNMP") {
        throw "Smoke failed: expected SNMP summary output in topology mode"
    }
}
finally {
    Remove-Item -Path $outputFile -ErrorAction SilentlyContinue
}

Write-Host "Smoke passed: topology mode prints SNMP summary." -ForegroundColor Green
