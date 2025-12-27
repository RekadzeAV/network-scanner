#!/bin/bash

# Скрипт для сборки сканера сети для разных платформ

echo "Сборка Network Scanner..."

# Создаем директорию для бинарников с датой сборки
BUILD_DATE=$(date +%Y-%m-%d)
RELEASE_DIR="dist/${BUILD_DATE}"
mkdir -p "${RELEASE_DIR}"
echo "📦 Бинарники будут сохранены в: ${RELEASE_DIR}/"
echo ""

# Текущая платформа
echo "Сборка для текущей платформы..."
go build -ldflags="-s -w" -o "${RELEASE_DIR}/network-scanner" ./cmd/network-scanner

# Linux 64-bit
echo "Сборка для Linux 64-bit..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "${RELEASE_DIR}/network-scanner-linux-amd64" ./cmd/network-scanner

# Windows 64-bit
echo "Сборка для Windows 64-bit..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "${RELEASE_DIR}/network-scanner-windows-amd64.exe" ./cmd/network-scanner

# macOS Intel
echo "Сборка для macOS Intel..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o "${RELEASE_DIR}/network-scanner-darwin-amd64" ./cmd/network-scanner

# macOS Apple Silicon
echo "Сборка для macOS Apple Silicon..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "${RELEASE_DIR}/network-scanner-darwin-arm64" ./cmd/network-scanner

echo ""
echo "✅ Сборка завершена! Бинарники находятся в директории ${RELEASE_DIR}/"

