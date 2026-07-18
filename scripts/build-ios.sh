#!/bin/bash
# Скрипт для сборки Network Scanner под iOS
#
# Требования:
# - macOS
# - Xcode с командной строкой
# - gomobile установлен: go install golang.org/x/mobile/cmd/gomobile@latest
# - gomobile init (первичная инициализация)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
OUTPUT_DIR="$SCRIPT_DIR/../release-build"

echo "🍎 Сборка Network Scanner для iOS..."

# Проверяем платформу
if [[ "$(uname)" != "Darwin" ]]; then
    echo "❌ Ошибка: iOS сборка доступна только на macOS"
    exit 1
fi

# Создаем директорию вывода
mkdir -p "$OUTPUT_DIR"

# Очищаем предыдущую сборку
rm -rf "$OUTPUT_DIR/network-scanner-ios"

# Сборка iOS framework
cd "$PROJECT_DIR"
gomobile bind -target=ios -o="NetworkScanner" \
    ./...

# Перемещаем артефакты
if [ -d "NetworkScanner.xcframework" ]; then
    mv "NetworkScanner.xcframework" "$OUTPUT_DIR/network-scanner-ios/"
    echo "✅ Framework: $OUTPUT_DIR/network-scanner-ios/NetworkScanner.xcframework"
fi

if [ -f "NetworkScanner.framework.zip" ]; then
    mv "NetworkScanner.framework.zip" "$OUTPUT_DIR/network-scanner-ios.zip"
    echo "✅ ZIP: $OUTPUT_DIR/network-scanner-ios.zip"
fi

echo "📦 Артефакты собраны в: $OUTPUT_DIR"
