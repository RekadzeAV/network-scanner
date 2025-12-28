#!/bin/bash

# Скрипт для проверки и настройки окружения для кросскомпиляции в Windows

echo "=========================================="
echo "Проверка окружения для сборки под Windows"
echo "=========================================="
echo ""

# Проверка Go
echo "[1/4] Проверка Go..."
if command -v go &> /dev/null; then
    GO_VERSION=$(go version)
    echo "✅ $GO_VERSION"
else
    echo "❌ Go не установлен"
    echo "   Установите: brew install go"
    exit 1
fi
echo ""

# Проверка CGO
echo "[2/4] Проверка CGO..."
CGO_ENABLED=$(go env CGO_ENABLED)
if [ "$CGO_ENABLED" = "1" ]; then
    echo "✅ CGO включен"
else
    echo "⚠️  CGO отключен, включаем..."
    export CGO_ENABLED=1
    echo "✅ CGO включен"
fi
echo ""

# Проверка mingw-w64
echo "[3/4] Проверка mingw-w64..."
if command -v x86_64-w64-mingw32-gcc &> /dev/null; then
    MINGW_VERSION=$(x86_64-w64-mingw32-gcc --version | head -n 1)
    echo "✅ mingw-w64 найден: $MINGW_VERSION"
    MINGW_INSTALLED=true
else
    echo "❌ mingw-w64 не найден"
    echo ""
    echo "Для установки выполните:"
    echo "  brew install mingw-w64"
    echo ""
    echo "После установки перезапустите скрипт."
    MINGW_INSTALLED=false
fi
echo ""

# Проверка зависимостей Go
echo "[4/4] Проверка зависимостей Go..."
cd "$(dirname "$0")/.."
if [ -f "go.mod" ]; then
    echo "📦 Загрузка зависимостей..."
    go mod download
    echo "✅ Зависимости готовы"
else
    echo "⚠️  go.mod не найден"
fi
echo ""

# Итоги
echo "=========================================="
if [ "$MINGW_INSTALLED" = true ]; then
    echo "✅ Окружение готово для сборки под Windows!"
    echo ""
    echo "Для сборки выполните:"
    echo "  ./scripts/build-windows.sh"
else
    echo "⚠️  Требуется установка mingw-w64"
    echo ""
    echo "Выполните:"
    echo "  brew install mingw-w64"
fi
echo "=========================================="

