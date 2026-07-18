#!/bin/bash
# Скрипт для сборки Mobile Scanner под Android
#
# Требования:
# - Android SDK (ANDROID_HOME)
# - JDK 17+
# - gomobile установлен: go install golang.org/x/mobile/cmd/gomobile@latest
# - android SDK установлен: go install golang.org/x/mobile/cmd/gobind@latest

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
OUTPUT_DIR="$SCRIPT_DIR/../release-build"

echo "📱 Сборка Network Scanner для Android..."

# Проверяем наличие ANDROID_HOME
if [ -z "$ANDROID_HOME" ] && [ -z "$ANDROID_SDK_ROOT" ]; then
    echo "❌ Ошибка: ANDROID_HOME не установлен"
    echo "Установите Android SDK и задайте переменную окружения"
    exit 1
fi

# Создаем директорию вывода
mkdir -p "$OUTPUT_DIR"

# Очищаем предыдущую сборку
rm -f "$OUTPUT_DIR/network-scanner-android.apk"

# Сборка APK
cd "$PROJECT_DIR"
gomobile bind -target=android -o="NetworkScanner" \
    -javapkg=network.scanner \
    -androidapi=21 \
    ./...

# Переименовываем APK
if [ -f "NetworkScanner.jar" ]; then
    mv NetworkScanner.jar "$OUTPUT_DIR/network-scanner-android.aar"
    echo "✅ Сборка завершена: $OUTPUT_DIR/network-scanner-android.aar"
fi

if [ -f "NetworkScanner-sources.jar" ]; then
    mv NetworkScanner-sources.jar "$OUTPUT_DIR/network-scanner-android-sources.jar"
    echo "✅ Источники: $OUTPUT_DIR/network-scanner-android-sources.jar"
fi

echo "📦 Артефакты собраны в: $OUTPUT_DIR"
