#!/bin/bash
# build-release.sh
# Script to build release version of network-scanner

set -e

VERSION="${1:-1.0.0}"
OUTPUT_DIR="${2:-build/release}"

echo "=== Network Scanner Release Build ==="
echo "Version: $VERSION"
echo "Output: $OUTPUT_DIR"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Run tests
echo ""
echo "Running tests..."
go test ./... -v
if [ $? -ne 0 ]; then
    echo "Tests failed! Aborting build."
    exit 1
fi
echo "Tests passed!"

# Build for Linux amd64
echo ""
echo "Building for Linux amd64..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$VERSION" -o "$OUTPUT_DIR/network-scanner-linux-amd64" ./cmd/network-scanner

# Build for Linux arm64
echo "Building for Linux arm64..."
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.Version=$VERSION" -o "$OUTPUT_DIR/network-scanner-linux-arm64" ./cmd/network-scanner

# Build for macOS amd64
echo "Building for macOS amd64..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$VERSION" -o "$OUTPUT_DIR/network-scanner-darwin-amd64" ./cmd/network-scanner

# Build for macOS arm64 (Apple Silicon)
echo "Building for macOS arm64..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.Version=$VERSION" -o "$OUTPUT_DIR/network-scanner-darwin-arm64" ./cmd/network-scanner

# Generate checksums
echo ""
echo "Generating checksums..."
cd "$OUTPUT_DIR"
sha256sum * > checksums.txt
cd ..

echo ""
echo "=== Build Complete ==="
echo "Artifacts:"
ls -lh "$OUTPUT_DIR"
