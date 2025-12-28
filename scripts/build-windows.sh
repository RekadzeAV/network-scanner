#!/bin/bash

# Скрипт для сборки Network Scanner для Windows (с macOS)
# Требует mingw-w64 для кросскомпиляции с CGO

set -e  # Остановка при ошибке

echo "=========================================="
echo "Сборка Network Scanner для Windows"
echo "=========================================="

# Проверка наличия Go
if ! command -v go &> /dev/null; then
    echo "❌ Ошибка: Go не установлен!"
    echo ""
    echo "Установите Go:"
    echo "1. Через Homebrew: brew install go"
    echo "2. Скачайте с официального сайта: https://go.dev/dl/"
    echo ""
    exit 1
fi

echo "✅ Go найден: $(go version)"
echo ""

# Проверка CGO
CGO_ENABLED=$(go env CGO_ENABLED)
if [ "$CGO_ENABLED" != "1" ]; then
    echo "⚠️  Предупреждение: CGO отключен. Включение CGO..."
    export CGO_ENABLED=1
fi
echo "✅ CGO включен: $CGO_ENABLED"
echo ""

# Проверка наличия mingw-w64
MINGW_GCC=""
if command -v x86_64-w64-mingw32-gcc &> /dev/null; then
    MINGW_GCC="x86_64-w64-mingw32-gcc"
    echo "✅ Найден mingw-w64: $MINGW_GCC"
elif command -v i686-w64-mingw32-gcc &> /dev/null; then
    MINGW_GCC="i686-w64-mingw32-gcc"
    echo "✅ Найден mingw-w64 (32-bit): $MINGW_GCC"
else
    echo "❌ Ошибка: mingw-w64 не найден!"
    echo ""
    echo "Установите mingw-w64 одним из способов:"
    echo ""
    echo "1. Через Homebrew (рекомендуется):"
    echo "   brew install mingw-w64"
    echo ""
    echo "2. Вручную:"
    echo "   - Скачайте с https://www.mingw-w64.org/downloads/"
    echo "   - Добавьте в PATH"
    echo ""
    echo "После установки перезапустите скрипт."
    exit 1
fi

echo ""

# Создаем директорию для бинарников с датой сборки
BUILD_DATE=$(date +%Y-%m-%d)
RELEASE_DIR="release/${BUILD_DATE}"
mkdir -p "${RELEASE_DIR}"
echo "📦 Бинарники будут сохранены в: ${RELEASE_DIR}/"
echo ""

# Установка зависимостей
echo "📦 Установка зависимостей..."
go mod download
go mod tidy
echo "✅ Зависимости установлены"
echo ""

# Настройка переменных окружения для Windows кросскомпиляции
export GOOS=windows
export GOARCH=amd64
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-g++
export CGO_ENABLED=1

echo "🔨 Сборка GUI версии для Windows 64-bit..."
go build -ldflags="-s -w" -o "${RELEASE_DIR}/network-scanner-gui-windows-amd64.exe" ./cmd/gui
echo "✅ Собрано: ${RELEASE_DIR}/network-scanner-gui-windows-amd64.exe"

# Сброс переменных окружения
unset GOOS
unset GOARCH
unset CC
unset CXX

echo ""
echo "=========================================="
echo "✅ Сборка завершена!"
echo "=========================================="
echo ""
echo "Собранные файлы находятся в директории ${RELEASE_DIR}/:"
ls -lh "${RELEASE_DIR}"/network-scanner-gui-windows-amd64.exe 2>/dev/null || echo "Файлы не найдены"
echo ""

