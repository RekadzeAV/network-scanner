Param()

$ErrorActionPreference = "Stop"

Write-Host "==> Checking Go toolchain"
$goVersion = go version
if (-not $goVersion) {
    throw "Go is not installed or not available in PATH. Install Go 1.24+ first."
}

Write-Host "==> Go version"
Write-Host $goVersion

Write-Host "==> Downloading module dependencies"
go mod download

Write-Host "==> Building CLI binary"
go build -o network-scanner.exe ./cmd/network-scanner

Write-Host "==> Running unit/integration tests"
go test ./...

Write-Host "Bootstrap completed successfully."
