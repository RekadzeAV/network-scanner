#!/bin/bash

# Скрипт для проверки сборки проекта

set -e

echo "=========================================="
echo "Проверка сборки Network Scanner"
echo "=========================================="

# Проверка наличия Go
if ! command -v go &> /dev/null; then
    echo "❌ Ошибка: Go не установлен!"
    echo ""
    echo "Установите Go одним из способов:"
    echo "1. Через Homebrew: brew install go"
    echo "2. Скачайте с https://go.dev/dl/"
    echo ""
    exit 1
fi

echo "✅ Go найден: $(go version)"
echo ""

# Переход в корень проекта
cd "$(dirname "$0")/.."

# Проверка структуры проекта
echo "📁 Проверка структуры проекта..."
if [ ! -d "cmd/network-scanner" ]; then
    echo "❌ Ошибка: cmd/network-scanner не найден"
    exit 1
fi

if [ ! -d "internal/scanner" ]; then
    echo "❌ Ошибка: internal/scanner не найден"
    exit 1
fi

if [ ! -d "internal/network" ]; then
    echo "❌ Ошибка: internal/network не найден"
    exit 1
fi

if [ ! -d "internal/display" ]; then
    echo "❌ Ошибка: internal/display не найден"
    exit 1
fi

echo "✅ Структура проекта корректна"
echo ""

# Проверка зависимостей
echo "📦 Проверка зависимостей..."
go mod download
go mod tidy
echo "✅ Зависимости проверены"
echo ""

# Проверка синтаксиса
echo "🔍 Проверка синтаксиса..."
if ! go build ./...; then
    echo "❌ Ошибка компиляции!"
    exit 1
fi
echo "✅ Синтаксис корректен"
echo ""

# Проверка линтера (если установлен)
if command -v golangci-lint &> /dev/null; then
    echo "🔍 Запуск линтера..."
    golangci-lint run ./... || echo "⚠️  Линтер нашел проблемы (не критично)"
    echo ""
fi

# Попытка сборки
echo "🔨 Попытка сборки..."
mkdir -p build/release
if go build -o build/release/network-scanner-test ./cmd/network-scanner; then
    echo "✅ Сборка успешна!"
    echo ""
    echo "Бинарник создан: build/release/network-scanner-test"
    echo ""
    echo "Для запуска:"
    echo "  ./build/release/network-scanner-test"
else
    echo "❌ Ошибка сборки!"
    exit 1
fi

echo "=========================================="
echo "✅ Проверка завершена успешно!"
echo "=========================================="




