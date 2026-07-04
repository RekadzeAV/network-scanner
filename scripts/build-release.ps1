#!/usr/bin/env pwsh
#Requires -Version 5.1

# build-release.ps1
# Скрипт для сборки релизной версии network-scanner

param(
    [string]$Version = "1.0.0",
    [string]$OutputDir = "build/release",
    [switch]$SkipTests
)

$ErrorActionPreference = "Stop"

Write-Host "=== Network Scanner Release Build ===" -ForegroundColor Cyan
Write-Host "Version: $Version" -ForegroundColor Yellow
Write-Host "Output: $OutputDir" -ForegroundColor Yellow

# Создаём директорию
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

# Запускаем тесты
if (-not $SkipTests) {
    Write-Host "`nRunning tests..." -ForegroundColor Green
    go test ./... -v
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Tests failed! Aborting build." -ForegroundColor Red
        exit 1
    }
    Write-Host "Tests passed!" -ForegroundColor Green
}

# Собираем для Windows
Write-Host "`nBuilding for Windows amd64..." -ForegroundColor Green
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$buildTime = Get-Date -Format 'yyyy-MM-dd HH:mm:ss'
$gitCommit = if ($env:GITHUB_SHA) { $env:GITHUB_SHA.Substring(0, 7) } else { "local" }
go build -ldflags="-s -w -X main.Version=$Version -X main.BuildTime=$buildTime -X main.GitCommit=$gitCommit" -o "$OutputDir/network-scanner-windows-amd64.exe" ./cmd/network-scanner
if ($LASTEXITCODE -ne 0) {
    Write-Host "Windows build failed!" -ForegroundColor Red
    exit 1
}

# Собираем для Linux
Write-Host "Building for Linux amd64..." -ForegroundColor Green
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -ldflags="-s -w -X main.Version=$Version" -o "$OutputDir/network-scanner-linux-amd64" ./cmd/network-scanner
if ($LASTEXITCODE -ne 0) {
    Write-Host "Linux build failed!" -ForegroundColor Red
    exit 1
}

# Собираем для macOS
Write-Host "Building for macOS amd64..." -ForegroundColor Green
$env:GOOS = "darwin"
$env:GOARCH = "amd64"
go build -ldflags="-s -w -X main.Version=$Version" -o "$OutputDir/network-scanner-darwin-amd64" ./cmd/network-scanner
if ($LASTEXITCODE -ne 0) {
    Write-Host "macOS build failed!" -ForegroundColor Red
    exit 1
}

# Создаём checksums
Write-Host "`nGenerating checksums..." -ForegroundColor Green
Set-Location $OutputDir
Get-ChildItem -File | ForEach-Object {
    $hash = Get-FileHash $_ -Algorithm SHA256
    "$($hash.Hash)  $($_.Name)"
} | Out-File -FilePath "checksums.txt" -Encoding utf8
Set-Location ..

Write-Host "`n=== Build Complete ===" -ForegroundColor Green
Write-Host "Artifacts:" -ForegroundColor Cyan
Get-ChildItem $OutputDir | Format-Table Name, Length
