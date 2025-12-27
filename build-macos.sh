#!/bin/bash

# Скрипт для сборки Network Scanner для macOS
# Поддерживает обе архитектуры: Intel (amd64) и Apple Silicon (arm64)

set -e  # Остановка при ошибке

echo "=========================================="
echo "Сборка Network Scanner для macOS"
echo "=========================================="

# Проверка наличия Go
if ! command -v go &> /dev/null; then
    echo "❌ Ошибка: Go не установлен!"
    echo ""
    echo "Установите Go одним из способов:"
    echo "1. Через Homebrew: brew install go"
    echo "2. Скачайте с официального сайта: https://go.dev/dl/"
    echo ""
    exit 1
fi

echo "✅ Go найден: $(go version)"
echo ""

# Создаем директорию для бинарников
mkdir -p dist

# Установка зависимостей
echo "📦 Установка зависимостей..."
go mod download
go mod tidy
echo "✅ Зависимости установлены"
echo ""

# Определяем текущую архитектуру
ARCH=$(uname -m)
echo "Текущая архитектура: $ARCH"
echo ""

# Сборка для текущей архитектуры
if [ "$ARCH" = "arm64" ]; then
    echo "🔨 Сборка для Apple Silicon (arm64)..."
    GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/network-scanner-darwin-arm64
    echo "✅ Собрано: dist/network-scanner-darwin-arm64"
elif [ "$ARCH" = "x86_64" ]; then
    echo "🔨 Сборка для Intel (amd64)..."
    GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/network-scanner-darwin-amd64
    echo "✅ Собрано: dist/network-scanner-darwin-amd64"
fi

# Попытка собрать для обеих архитектур (если возможно)
echo ""
echo "🔨 Попытка собрать универсальный бинарник (universal binary)..."

# Проверяем наличие lipo (для создания universal binary)
if command -v lipo &> /dev/null; then
    # Собираем для обеих архитектур
    echo "Сборка для Intel (amd64)..."
    GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/network-scanner-darwin-amd64-temp
    
    echo "Сборка для Apple Silicon (arm64)..."
    GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/network-scanner-darwin-arm64-temp
    
    # Создаем universal binary
    echo "Создание universal binary..."
    lipo -create \
        dist/network-scanner-darwin-amd64-temp \
        dist/network-scanner-darwin-arm64-temp \
        -output dist/network-scanner-darwin-universal
    
    # Удаляем временные файлы
    rm dist/network-scanner-darwin-amd64-temp
    rm dist/network-scanner-darwin-arm64-temp
    
    echo "✅ Создан универсальный бинарник: dist/network-scanner-darwin-universal"
else
    echo "⚠️  lipo не найден, пропускаем создание universal binary"
    echo "   (это нормально, если вы не используете Xcode Command Line Tools)"
fi

echo ""
echo "=========================================="
echo "✅ Сборка завершена!"
echo "=========================================="
echo ""
echo "Собранные файлы находятся в директории dist/:"
ls -lh dist/network-scanner-darwin* 2>/dev/null || echo "Файлы не найдены"
echo ""
echo "Для запуска:"
echo "  ./dist/network-scanner-darwin-<arch>"
echo ""

