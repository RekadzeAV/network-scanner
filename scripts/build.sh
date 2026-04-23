#!/bin/bash

# Скрипт для сборки сканера сети для разных платформ

echo "Сборка Network Scanner..."

# Создаем директорию для бинарников с датой сборки и номером
BUILD_DATE=$(date +%Y-%m-%d)
BUILD_NUM=1
RELEASE_DIR="build/release/${BUILD_DATE}-${BUILD_NUM}"

# Находим следующий доступный номер сборки
while [ -d "${RELEASE_DIR}" ]; do
    BUILD_NUM=$((BUILD_NUM + 1))
    RELEASE_DIR="build/release/${BUILD_DATE}-${BUILD_NUM}"
done

mkdir -p "${RELEASE_DIR}"
echo "📦 Бинарники будут сохранены в: ${RELEASE_DIR}/ (сборка #${BUILD_NUM})"
echo ""

# Текущая платформа
echo "Сборка CLI версии для текущей платформы..."
go build -ldflags="-s -w" -o "${RELEASE_DIR}/network-scanner" ./cmd/network-scanner

echo "Сборка GUI версии для текущей платформы..."
# Fyne требует CGO для GUI приложений
CGO_ENABLED=1 go build -ldflags="-s -w" -o "${RELEASE_DIR}/network-scanner-gui" ./cmd/gui

# Linux 64-bit
echo "Сборка CLI версии для Linux 64-bit..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "${RELEASE_DIR}/network-scanner-linux-amd64" ./cmd/network-scanner

echo "Сборка GUI версии для Linux 64-bit..."
# Fyne требует CGO для GUI приложений
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "${RELEASE_DIR}/network-scanner-gui-linux-amd64" ./cmd/gui

# Windows 64-bit
echo "Сборка CLI версии для Windows 64-bit..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "${RELEASE_DIR}/network-scanner-windows-amd64.exe" ./cmd/network-scanner

# Windows 64-bit GUI (требует mingw-w64 для CGO)
echo "Сборка GUI версии для Windows 64-bit..."
if command -v x86_64-w64-mingw32-gcc &> /dev/null; then
    GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ CGO_ENABLED=1 go build -ldflags="-s -w -H windowsgui" -o "${RELEASE_DIR}/network-scanner-gui-windows-amd64.exe" ./cmd/gui
    echo "✅ Собрано: ${RELEASE_DIR}/network-scanner-gui-windows-amd64.exe"
else
    echo "⚠️  mingw-w64 не найден, пропускаем сборку GUI для Windows"
    echo "   Установите: brew install mingw-w64"
    echo "   Или используйте скрипт: ./scripts/build-windows.sh"
fi

# macOS Intel
echo "Сборка CLI версии для macOS Intel..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o "${RELEASE_DIR}/network-scanner-darwin-amd64" ./cmd/network-scanner

echo "Сборка GUI версии для macOS Intel..."
# Fyne требует CGO для GUI приложений
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o "${RELEASE_DIR}/network-scanner-gui-darwin-amd64" ./cmd/gui

# macOS Apple Silicon
echo "Сборка CLI версии для macOS Apple Silicon..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "${RELEASE_DIR}/network-scanner-darwin-arm64" ./cmd/network-scanner

echo "Сборка GUI версии для macOS Apple Silicon..."
# Fyne требует CGO для GUI приложений
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "${RELEASE_DIR}/network-scanner-gui-darwin-arm64" ./cmd/gui

# Копируем инструкцию по эксплуатации в папку релиза
if [ -f "Инструкция по эксплуатации.md" ]; then
    cp "Инструкция по эксплуатации.md" "${RELEASE_DIR}/"
    echo "✅ Инструкция по эксплуатации скопирована в ${RELEASE_DIR}/"
else
    echo "⚠️  Файл 'Инструкция по эксплуатации.md' не найден в корне проекта"
fi

echo ""
echo "✅ Сборка завершена! Бинарники находятся в директории ${RELEASE_DIR}/"

