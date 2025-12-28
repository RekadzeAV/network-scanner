# Настройка кросскомпиляции для Windows на macOS

## Обзор

Данный документ описывает настройку окружения для сборки Network Scanner под Windows на macOS.

## Требования

### Обязательные компоненты

1. **Go 1.21+** - основной компилятор
2. **mingw-w64** - кросс-компилятор для Windows (требуется для CGO)

### Почему нужен mingw-w64?

Network Scanner использует GUI framework Fyne, который требует CGO. Для кросскомпиляции с CGO необходим соответствующий C компилятор для целевой платформы.

## Установка mingw-w64

### Вариант 1: Через Homebrew (рекомендуется)

```bash
# Установка Homebrew (если еще не установлен)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Установка mingw-w64
brew install mingw-w64
```

После установки проверьте:

```bash
x86_64-w64-mingw32-gcc --version
```

### Вариант 2: Вручную

1. Скачайте mingw-w64 с официального сайта: https://www.mingw-w64.org/downloads/
2. Распакуйте архив
3. Добавьте путь к `bin` в переменную PATH:
   ```bash
   export PATH="/path/to/mingw64/bin:$PATH"
   ```

## Настройка окружения

### Автоматическая настройка (рекомендуется)

Используйте скрипт сборки, который автоматически настраивает окружение:

```bash
chmod +x scripts/build-windows.sh
./scripts/build-windows.sh
```

### Ручная настройка

Если нужно настроить вручную, установите следующие переменные окружения:

```bash
export GOOS=windows
export GOARCH=amd64
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-g++
export CGO_ENABLED=1
```

## Проверка готовности

Выполните следующие команды для проверки:

```bash
# Проверка Go
go version
# Должно показать: go version go1.21.x или выше

# Проверка CGO
go env CGO_ENABLED
# Должно быть: CGO_ENABLED="1"

# Проверка mingw-w64
x86_64-w64-mingw32-gcc --version
# Должно показать версию GCC

# Проверка зависимостей
cd "/Users/rav/Documents/Сканер локальной сети"
go mod download
```

## Сборка

### Использование скрипта (рекомендуется)

```bash
./scripts/build-windows.sh
```

Скрипт автоматически:
- Проверит наличие всех необходимых инструментов
- Установит зависимости
- Настроит переменные окружения
- Соберет Windows версию
- Сохранит бинарник в `release/YYYY-MM-DD/`

### Ручная сборка

```bash
# Установка зависимостей
go mod download

# Настройка окружения
export GOOS=windows
export GOARCH=amd64
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-g++
export CGO_ENABLED=1

# Сборка GUI версии
go build -ldflags="-s -w" -o network-scanner-gui-windows-amd64.exe ./cmd/gui
```

## Тестирование

После сборки бинарник можно протестировать:

1. **Перенести на Windows машину** (через сеть, USB, etc.)
2. **Запустить**: `network-scanner-gui-windows-amd64.exe`

### Проверка бинарника на macOS

Можно проверить тип файла:

```bash
file release/YYYY-MM-DD/network-scanner-gui-windows-amd64.exe
```

Должно показать: `PE32+ executable (console) x86-64`

## Устранение проблем

### Ошибка: "x86_64-w64-mingw32-gcc: command not found"

**Причина**: mingw-w64 не установлен или не в PATH

**Решение**:
```bash
# Установить через Homebrew
brew install mingw-w64

# Проверить установку
which x86_64-w64-mingw32-gcc
```

### Ошибка: "CGO_ENABLED=0"

**Причина**: CGO отключен

**Решение**:
```bash
export CGO_ENABLED=1
# Или проверить глобальные настройки
go env CGO_ENABLED
```

### Ошибка: "cgo: C compiler 'x86_64-w64-mingw32-gcc' not found"

**Причина**: Компилятор не найден в PATH

**Решение**:
```bash
# Найти путь к mingw-w64
brew --prefix mingw-w64

# Добавить в PATH (пример)
export PATH="/opt/homebrew/opt/mingw-w64/bin:$PATH"

# Или создать симлинк
ln -s /opt/homebrew/opt/mingw-w64/bin/x86_64-w64-mingw32-gcc /usr/local/bin/
```

### Ошибки компиляции CGO

**Причина**: Недостающие системные библиотеки

**Решение**:
- Убедитесь, что mingw-w64 установлен полностью
- Проверьте наличие заголовочных файлов: `find $(brew --prefix mingw-w64) -name "*.h" | head`

## Альтернативные методы

### Использование Docker

Можно использовать Docker контейнер с предустановленными инструментами:

```dockerfile
FROM golang:1.21

RUN apt-get update && apt-get install -y \
    gcc-mingw-w64-x86-64 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY . .
RUN GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 go build ./cmd/gui
```

### Использование CI/CD

Настройте автоматическую сборку через GitHub Actions с Windows runner - сборка будет нативной.

## Дополнительные ресурсы

- [Go CGO Documentation](https://pkg.go.dev/cmd/cgo)
- [Go Cross Compilation](https://go.dev/doc/install/source#crosscompile)
- [mingw-w64 Official Site](https://www.mingw-w64.org/)
- [Fyne Cross Compilation](https://developer.fyne.io/started/cross-compiling)

