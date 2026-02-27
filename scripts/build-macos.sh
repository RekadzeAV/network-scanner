#!/bin/bash

# Скрипт для сборки Network Scanner для macOS
# Поддерживает обе архитектуры: Intel (amd64) и Apple Silicon (arm64)

# Не используем set -e, чтобы скрипт продолжал работу при ошибках сборки отдельных версий

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

# Создаем директорию для бинарников с датой сборки и номером
BUILD_DATE=$(date +%Y-%m-%d)
BUILD_NUM=1
RELEASE_DIR="release/${BUILD_DATE}-${BUILD_NUM}"

# Находим следующий доступный номер сборки
while [ -d "${RELEASE_DIR}" ]; do
    BUILD_NUM=$((BUILD_NUM + 1))
    RELEASE_DIR="release/${BUILD_DATE}-${BUILD_NUM}"
done

mkdir -p "${RELEASE_DIR}"
echo "📦 Бинарники будут сохранены в: ${RELEASE_DIR}/ (сборка #${BUILD_NUM})"
echo ""

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

# Флаги для отслеживания успешных сборок
BUILT_CLI_AMD64=false
BUILT_CLI_ARM64=false
BUILT_GUI_AMD64=false
BUILT_GUI_ARM64=false

# Сборка для текущей архитектуры
if [ "$ARCH" = "arm64" ]; then
    echo "🔨 Сборка CLI версии для Apple Silicon (arm64)..."
    if GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "${RELEASE_DIR}/network-scanner-darwin-arm64" ./cmd/network-scanner; then
        echo "✅ Собрано: ${RELEASE_DIR}/network-scanner-darwin-arm64"
        BUILT_CLI_ARM64=true
    else
        echo "❌ Ошибка сборки CLI для arm64"
    fi
    
    echo "🔨 Сборка GUI версии для Apple Silicon (arm64)..."
    # Fyne требует CGO для GUI приложений
    if CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "${RELEASE_DIR}/network-scanner-gui-darwin-arm64" ./cmd/gui; then
        echo "✅ Собрано: ${RELEASE_DIR}/network-scanner-gui-darwin-arm64"
        BUILT_GUI_ARM64=true
    else
        echo "❌ Ошибка сборки GUI для arm64"
        rm -f "${RELEASE_DIR}/network-scanner-gui-darwin-arm64"
    fi
elif [ "$ARCH" = "x86_64" ]; then
    echo "🔨 Сборка CLI версии для Intel (amd64)..."
    if GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o "${RELEASE_DIR}/network-scanner-darwin-amd64" ./cmd/network-scanner; then
        echo "✅ Собрано: ${RELEASE_DIR}/network-scanner-darwin-amd64"
        BUILT_CLI_AMD64=true
    else
        echo "❌ Ошибка сборки CLI для amd64"
    fi
    
    echo "🔨 Сборка GUI версии для Intel (amd64)..."
    # Fyne требует CGO для GUI приложений
    if CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o "${RELEASE_DIR}/network-scanner-gui-darwin-amd64" ./cmd/gui; then
        echo "✅ Собрано: ${RELEASE_DIR}/network-scanner-gui-darwin-amd64"
        BUILT_GUI_AMD64=true
    else
        echo "❌ Ошибка сборки GUI для amd64"
        rm -f "${RELEASE_DIR}/network-scanner-gui-darwin-amd64"
    fi
fi

# Попытка собрать для обеих архитектур и создать universal binary
echo ""
echo "🔨 Попытка собрать универсальные бинарники (universal binary)..."

# Проверяем наличие lipo (для создания universal binary)
if command -v lipo &> /dev/null; then
    # Создаем временную директорию для промежуточных файлов
    TEMP_DIR=$(mktemp -d)
    
    # Собираем CLI версию для обеих архитектур (если еще не собраны)
    if [ "$BUILT_CLI_AMD64" = false ]; then
        echo "Сборка CLI версии для Intel (amd64)..."
        if GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o "${TEMP_DIR}/network-scanner-darwin-amd64-temp" ./cmd/network-scanner; then
            BUILT_CLI_AMD64=true
        fi
    else
        if [ -f "${RELEASE_DIR}/network-scanner-darwin-amd64" ]; then
            cp "${RELEASE_DIR}/network-scanner-darwin-amd64" "${TEMP_DIR}/network-scanner-darwin-amd64-temp"
        else
            BUILT_CLI_AMD64=false
        fi
    fi
    
    if [ "$BUILT_CLI_ARM64" = false ]; then
        echo "Сборка CLI версии для Apple Silicon (arm64)..."
        if GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "${TEMP_DIR}/network-scanner-darwin-arm64-temp" ./cmd/network-scanner; then
            BUILT_CLI_ARM64=true
        fi
    else
        if [ -f "${RELEASE_DIR}/network-scanner-darwin-arm64" ]; then
            cp "${RELEASE_DIR}/network-scanner-darwin-arm64" "${TEMP_DIR}/network-scanner-darwin-arm64-temp"
        else
            BUILT_CLI_ARM64=false
        fi
    fi
    
    # Создаем universal binary для CLI
    if [ "$BUILT_CLI_AMD64" = true ] && [ "$BUILT_CLI_ARM64" = true ]; then
        echo "Создание universal binary для CLI..."
        if lipo -create \
            "${TEMP_DIR}/network-scanner-darwin-amd64-temp" \
            "${TEMP_DIR}/network-scanner-darwin-arm64-temp" \
            -output "${RELEASE_DIR}/network-scanner-darwin-universal" 2>&1; then
            echo "✅ Создан универсальный бинарник CLI: ${RELEASE_DIR}/network-scanner-darwin-universal"
        else
            echo "⚠️  Не удалось создать universal binary для CLI"
            rm -f "${RELEASE_DIR}/network-scanner-darwin-universal"
        fi
    fi
    
    # Собираем GUI версию для обеих архитектур (если еще не собраны)
    if [ "$BUILT_GUI_AMD64" = false ]; then
        echo "Сборка GUI версии для Intel (amd64)..."
        if CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o "${TEMP_DIR}/network-scanner-gui-darwin-amd64-temp" ./cmd/gui 2>&1; then
            BUILT_GUI_AMD64=true
        fi
    else
        if [ -f "${RELEASE_DIR}/network-scanner-gui-darwin-amd64" ]; then
            cp "${RELEASE_DIR}/network-scanner-gui-darwin-amd64" "${TEMP_DIR}/network-scanner-gui-darwin-amd64-temp"
        else
            BUILT_GUI_AMD64=false
        fi
    fi
    
    if [ "$BUILT_GUI_ARM64" = false ]; then
        echo "Сборка GUI версии для Apple Silicon (arm64)..."
        if CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "${TEMP_DIR}/network-scanner-gui-darwin-arm64-temp" ./cmd/gui 2>&1; then
            BUILT_GUI_ARM64=true
        fi
    else
        if [ -f "${RELEASE_DIR}/network-scanner-gui-darwin-arm64" ]; then
            cp "${RELEASE_DIR}/network-scanner-gui-darwin-arm64" "${TEMP_DIR}/network-scanner-gui-darwin-arm64-temp"
        else
            BUILT_GUI_ARM64=false
        fi
    fi
    
    # Создаем universal binary для GUI
    if [ "$BUILT_GUI_AMD64" = true ] && [ "$BUILT_GUI_ARM64" = true ]; then
        echo "Создание universal binary для GUI..."
        if lipo -create \
            "${TEMP_DIR}/network-scanner-gui-darwin-amd64-temp" \
            "${TEMP_DIR}/network-scanner-gui-darwin-arm64-temp" \
            -output "${RELEASE_DIR}/network-scanner-gui-darwin-universal" 2>&1; then
            echo "✅ Создан универсальный бинарник GUI: ${RELEASE_DIR}/network-scanner-gui-darwin-universal"
        else
            echo "⚠️  Не удалось создать universal binary для GUI"
            rm -f "${RELEASE_DIR}/network-scanner-gui-darwin-universal"
        fi
    fi
    
    # Удаляем временные файлы
    rm -rf "${TEMP_DIR}"
else
    echo "⚠️  lipo не найден, пропускаем создание universal binary"
    echo "   (это нормально, если вы не используете Xcode Command Line Tools)"
fi

# Копируем инструкцию по эксплуатации в папку релиза
if [ -f "Инструкция по эксплуатации.md" ]; then
    cp "Инструкция по эксплуатации.md" "${RELEASE_DIR}/"
    echo "✅ Инструкция по эксплуатации скопирована в ${RELEASE_DIR}/"
else
    echo "⚠️  Файл 'Инструкция по эксплуатации.md' не найден в корне проекта"
fi

echo ""
echo "=========================================="
echo "✅ Сборка завершена!"
echo "=========================================="
echo ""
echo "Собранные файлы находятся в директории ${RELEASE_DIR}/:"
echo ""
echo "CLI версии:"
ls -lh "${RELEASE_DIR}"/network-scanner-darwin-* 2>/dev/null | grep -v "gui" || echo "  (нет файлов)"
echo ""
echo "GUI версии:"
ls -lh "${RELEASE_DIR}"/network-scanner-gui-darwin-* 2>/dev/null || echo "  (нет файлов)"
echo ""
echo "Универсальные бинарники:"
ls -lh "${RELEASE_DIR}"/network-scanner*-universal 2>/dev/null || echo "  (нет файлов)"
echo ""
echo "Для запуска:"
echo "  CLI: ./${RELEASE_DIR}/network-scanner-darwin-<arch>"
echo "  GUI: ./${RELEASE_DIR}/network-scanner-gui-darwin-<arch>"
echo ""

