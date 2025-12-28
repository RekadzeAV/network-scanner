#!/bin/bash

# Скрипт для запуска GUI приложения без терминала на macOS
# Использует команду 'open' для запуска приложения как GUI приложения

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Ищем скомпилированное GUI приложение
GUI_APP=""

# Проверяем последнюю дату сборки в dist
LATEST_DATE=$(ls -t "${SCRIPT_DIR}/dist" 2>/dev/null | head -1)

if [ -n "$LATEST_DATE" ]; then
    # Проверяем универсальный бинарник
    if [ -f "${SCRIPT_DIR}/dist/${LATEST_DATE}/gui-darwin-universal" ]; then
        GUI_APP="${SCRIPT_DIR}/dist/${LATEST_DATE}/gui-darwin-universal"
    # Проверяем по архитектуре
    elif [ "$(uname -m)" = "arm64" ] && [ -f "${SCRIPT_DIR}/dist/${LATEST_DATE}/gui-darwin-arm64" ]; then
        GUI_APP="${SCRIPT_DIR}/dist/${LATEST_DATE}/gui-darwin-arm64"
    elif [ "$(uname -m)" = "x86_64" ] && [ -f "${SCRIPT_DIR}/dist/${LATEST_DATE}/gui-darwin-amd64" ]; then
        GUI_APP="${SCRIPT_DIR}/dist/${LATEST_DATE}/gui-darwin-amd64"
    fi
fi

# Если не нашли в dist, проверяем корневую директорию
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

