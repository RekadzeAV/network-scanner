#!/bin/bash

# Скрипт для сборки сканера сети для разных платформ

echo "Сборка Network Scanner..."

# Создаем директорию для бинарников
mkdir -p dist

# Текущая платформа
echo "Сборка для текущей платформы..."
go build -o dist/network-scanner ./cmd/network-scanner

# Linux 64-bit
echo "Сборка для Linux 64-bit..."
GOOS=linux GOARCH=amd64 go build -o dist/network-scanner-linux-amd64 ./cmd/network-scanner

# Windows 64-bit
echo "Сборка для Windows 64-bit..."
GOOS=windows GOARCH=amd64 go build -o dist/network-scanner-windows-amd64.exe ./cmd/network-scanner

# macOS Intel
echo "Сборка для macOS Intel..."
GOOS=darwin GOARCH=amd64 go build -o dist/network-scanner-darwin-amd64 ./cmd/network-scanner

# macOS Apple Silicon
echo "Сборка для macOS Apple Silicon..."
GOOS=darwin GOARCH=arm64 go build -o dist/network-scanner-darwin-arm64 ./cmd/network-scanner

echo "Сборка завершена! Бинарники находятся в директории dist/"

