#!/bin/bash

# Скрипт для запуска GUI приложения без терминала на macOS
# Использует команду 'open' для запуска приложения как GUI приложения

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Ищем скомпилированное GUI приложение
GUI_APP=""

# Релизные скрипты кладут GUI в build/release/<YYYY-MM-DD-N>/ (раньше использовался dist/)
RELEASE_BASE=""
if [ -d "${SCRIPT_DIR}/build/release" ]; then
    RELEASE_BASE="${SCRIPT_DIR}/build/release"
elif [ -d "${SCRIPT_DIR}/dist" ]; then
    RELEASE_BASE="${SCRIPT_DIR}/dist"
fi

LATEST_DATE=""
if [ -n "$RELEASE_BASE" ]; then
    LATEST_DATE=$(ls -t "$RELEASE_BASE" 2>/dev/null | head -1)
fi

if [ -n "$LATEST_DATE" ]; then
    R="${RELEASE_BASE}/${LATEST_DATE}"
    if [ -f "${R}/gui-darwin-universal" ]; then
        GUI_APP="${R}/gui-darwin-universal"
    elif [ "$(uname -m)" = "arm64" ] && [ -f "${R}/gui-darwin-arm64" ]; then
        GUI_APP="${R}/gui-darwin-arm64"
    elif [ "$(uname -m)" = "x86_64" ] && [ -f "${R}/gui-darwin-amd64" ]; then
        GUI_APP="${R}/gui-darwin-amd64"
    fi
fi

# Если не нашли в build/release (или dist), проверяем корневую директорию
if [ -z "$GUI_APP" ]; then
    if [ -f "${SCRIPT_DIR}/gui" ]; then
        GUI_APP="${SCRIPT_DIR}/gui"
    elif [ -f "${SCRIPT_DIR}/gui-darwin-universal" ]; then
        GUI_APP="${SCRIPT_DIR}/gui-darwin-universal"
    elif [ -f "${SCRIPT_DIR}/gui-darwin-arm64" ]; then
        GUI_APP="${SCRIPT_DIR}/gui-darwin-arm64"
    elif [ -f "${SCRIPT_DIR}/gui-darwin-amd64" ]; then
        GUI_APP="${SCRIPT_DIR}/gui-darwin-amd64"
    fi
fi

if [ -z "$GUI_APP" ] || [ ! -f "$GUI_APP" ]; then
    echo "❌ GUI приложение не найдено!"
    echo ""
    echo "Сначала соберите приложение:"
    echo "  ./scripts/build-gui-release.sh"
    echo ""
    echo "Или соберите вручную:"
    echo "  go build -o gui ./cmd/gui"
    exit 1
fi

# На macOS используем 'open' для запуска без терминала
if [[ "$OSTYPE" == "darwin"* ]]; then
    # Используем open для запуска приложения как GUI приложения
    # Это предотвратит появление терминала
    open "$GUI_APP"
else
    # На других системах просто запускаем
    "$GUI_APP"
fi

