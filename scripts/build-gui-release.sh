#!/bin/bash

# Скрипт для сборки GUI приложения Network Scanner для macOS
# Поддерживает обе архитектуры: Intel (amd64) и Apple Silicon (arm64)

set -e  # Остановка при ошибке

echo "=========================================="
echo "Сборка GUI приложения Network Scanner для macOS"
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

# Сборка для обеих архитектур macOS
# Fyne требует CGO для GUI приложений
BUILT_AMD64=false
BUILT_ARM64=false

echo "🔨 Сборка для Intel (amd64)..."
if CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o "${RELEASE_DIR}/gui-darwin-amd64" ./cmd/gui 2>&1; then
    echo "✅ Собрано: ${RELEASE_DIR}/gui-darwin-amd64"
    BUILT_AMD64=true
else
    echo "⚠️  Не удалось собрать для amd64"
    rm -f "${RELEASE_DIR}/gui-darwin-amd64"
fi

echo ""
echo "🔨 Сборка для Apple Silicon (arm64)..."
if CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "${RELEASE_DIR}/gui-darwin-arm64" ./cmd/gui 2>&1; then
    echo "✅ Собрано: ${RELEASE_DIR}/gui-darwin-arm64"
    BUILT_ARM64=true
else
    echo "⚠️  Не удалось собрать для arm64 (кросс-компиляция может быть недоступна)"
    echo "   Для сборки arm64 версии используйте Mac с Apple Silicon"
    rm -f "${RELEASE_DIR}/gui-darwin-arm64"
fi

# Попытка собрать универсальный бинарник (если обе версии собрались)
if [ "$BUILT_AMD64" = true ] && [ "$BUILT_ARM64" = true ]; then
    echo ""
    echo "🔨 Попытка собрать универсальный бинарник (universal binary)..."
    
    # Проверяем наличие lipo (для создания universal binary)
    if command -v lipo &> /dev/null; then
        # Создаем universal binary
        echo "Создание universal binary..."
        if lipo -create \
            "${RELEASE_DIR}/gui-darwin-amd64" \
            "${RELEASE_DIR}/gui-darwin-arm64" \
            -output "${RELEASE_DIR}/gui-darwin-universal" 2>&1; then
            echo "✅ Создан универсальный бинарник: ${RELEASE_DIR}/gui-darwin-universal"
        else
            echo "⚠️  Не удалось создать universal binary"
            rm -f "${RELEASE_DIR}/gui-darwin-universal"
        fi
    else
        echo "⚠️  lipo не найден, пропускаем создание universal binary"
        echo "   (это нормально, если вы не используете Xcode Command Line Tools)"
    fi
elif [ "$BUILT_AMD64" = false ] && [ "$BUILT_ARM64" = false ]; then
    echo ""
    echo "❌ Ошибка: Не удалось собрать ни одну версию!"
    exit 1
fi

# Создаем README для релиза
echo ""
echo "📝 Создание README для релиза..."
cat > "${RELEASE_DIR}/README.md" << EOF
# Network Scanner GUI Release ${BUILD_DATE}-${BUILD_NUM}

## Собранные файлы

### GUI приложение

EOF

if [ "$BUILT_AMD64" = true ]; then
    echo "- **gui-darwin-amd64** - macOS Intel (x86_64)" >> "${RELEASE_DIR}/README.md"
fi

if [ "$BUILT_ARM64" = true ]; then
    echo "- **gui-darwin-arm64** - macOS Apple Silicon (ARM64)" >> "${RELEASE_DIR}/README.md"
fi

if [ -f "${RELEASE_DIR}/gui-darwin-universal" ]; then
    echo "- **gui-darwin-universal** - macOS Universal Binary (Intel + Apple Silicon)" >> "${RELEASE_DIR}/README.md"
fi

cat >> "${RELEASE_DIR}/README.md" << EOF

## Использование

### GUI приложение

\`\`\`bash
EOF

if [ "$BUILT_AMD64" = true ]; then
    echo "# macOS Intel" >> "${RELEASE_DIR}/README.md"
    echo "./gui-darwin-amd64" >> "${RELEASE_DIR}/README.md"
    echo "" >> "${RELEASE_DIR}/README.md"
fi

if [ "$BUILT_ARM64" = true ]; then
    echo "# macOS Apple Silicon" >> "${RELEASE_DIR}/README.md"
    echo "./gui-darwin-arm64" >> "${RELEASE_DIR}/README.md"
    echo "" >> "${RELEASE_DIR}/README.md"
fi

if [ -f "${RELEASE_DIR}/gui-darwin-universal" ]; then
    echo "# macOS Universal (любая архитектура)" >> "${RELEASE_DIR}/README.md"
    echo "./gui-darwin-universal" >> "${RELEASE_DIR}/README.md"
    echo "" >> "${RELEASE_DIR}/README.md"
fi

cat >> "${RELEASE_DIR}/README.md" << EOF
\`\`\`

## Размеры файлов

- GUI: ~24-25 MB

## Примечания

### Windows и Linux версии

Windows и Linux GUI версии не включены в этот релиз, так как требуют сборки на соответствующих системах из-за зависимостей от CGO и системных библиотек. Для сборки используйте:

**Windows:**
\`\`\`bash
# На Windows системе
go build -o gui-windows-amd64.exe ./cmd/gui
\`\`\`

**Linux:**
\`\`\`bash
# На Linux системе
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o gui-linux-amd64 ./cmd/gui
\`\`\`
EOF

echo "✅ README создан"
echo ""

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
ls -lh "${RELEASE_DIR}"/gui-darwin* 2>/dev/null || echo "Файлы не найдены"
echo ""
echo "Для запуска:"
echo "  ./${RELEASE_DIR}/gui-darwin-<arch>"
echo ""

