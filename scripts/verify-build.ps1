#!/usr/bin/env pwsh
#Requires -Version 5.1

# verify-build.ps1
# Verifies that build artifacts exist and have the expected extension.
# On Windows executables MUST carry the .exe suffix, otherwise CLI/GUI cannot
# be launched by double-click and the release layout is broken.
#
# Usage:
#   powershell -ExecutionPolicy Bypass -File scripts\verify-build.ps1
#   powershell -ExecutionPolicy Bypass -File scripts\verify-build.ps1 -Dir build
#
# Exit code: 0 on success, 1 if any artifact is missing or has wrong extension.

param(
    [string]$Dir = "build",
    [string]$ExeSuffix = ".exe"
)

$ErrorActionPreference = "Stop"

$expected = @(
    "network-scanner$ExeSuffix",
    "network-scanner-gui$ExeSuffix"
)

$isWindowsHost = ($env:OS -eq "Windows_NT") -or $IsWindows

Write-Host "=== Verify build artifacts ===" -ForegroundColor Cyan
Write-Host "Directory: $Dir" -ForegroundColor Yellow
Write-Host "Expected suffix: '$ExeSuffix'" -ForegroundColor Yellow

if (-not (Test-Path $Dir)) {
    Write-Host "[FAIL] Build directory not found: $Dir" -ForegroundColor Red
    exit 1
}

$failed = $false
$verified = @()

foreach ($name in $expected) {
    $path = Join-Path $Dir $name
    if (-not (Test-Path $path)) {
        Write-Host "[FAIL] Missing artifact: $path" -ForegroundColor Red
        $failed = $true
        continue
    }

    $item = Get-Item $path
    if ($item.Length -le 0) {
        Write-Host "[FAIL] Empty artifact: $path" -ForegroundColor Red
        $failed = $true
        continue
    }

    if ($isWindowsHost -and ($ExeSuffix -eq ".exe")) {
        $ext = [System.IO.Path]::GetExtension($item.Name)
        if ($ext -ne ".exe") {
            Write-Host "[FAIL] Artifact without .exe extension: $path" -ForegroundColor Red
            $failed = $true
            continue
        }
    }

    # Detect stray extensionless duplicates, e.g. from 'go build -o build/network-scanner'.
    $withoutExt = Join-Path $Dir ([System.IO.Path]::GetFileNameWithoutExtension($item.Name))
    if (Test-Path $withoutExt -PathType Leaf) {
        Write-Host "[WARN] Found stray file without extension: $withoutExt" -ForegroundColor Yellow
        Write-Host "       Remove it to avoid confusion." -ForegroundColor Yellow
    }

    Write-Host ("[OK]   {0} ({1:N0} bytes)" -f $item.Name, $item.Length) -ForegroundColor Green
    $verified += $item.Name
}

if ($failed) {
    Write-Host ""
    Write-Host "Verification FAILED. Rebuild with: make build" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host ("Verification OK: {0}" -f ($verified -join ", ")) -ForegroundColor Green
exit 0
