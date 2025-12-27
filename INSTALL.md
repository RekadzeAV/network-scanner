# Инструкция по установке и сборке для macOS

## Установка Go

Если Go еще не установлен, выполните один из вариантов:

### Вариант 1: Через Homebrew (рекомендуется)

```bash
# Установка Homebrew (если еще не установлен)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Установка Go
brew install go
```

### Вариант 2: Прямая установка

1. Перейдите на https://go.dev/dl/
2. Скачайте установщик для macOS
3. Запустите установщик и следуйте инструкциям

### Проверка установки

```bash
go version
```

Должно вывести что-то вроде: `go version go1.21.x darwin/arm64` или `go version go1.21.x darwin/amd64`

## Сборка приложения

### Быстрая сборка

```bash
# Перейдите в директорию проекта
cd "Сканер локальной сети"

# Запустите скрипт сборки для macOS
./build-macos.sh
```

### Ручная сборка

```bash
# Установка зависимостей
go mod download

# Сборка для Apple Silicon (M1/M2/M3)
GOOS=darwin GOARCH=arm64 go build -o network-scanner-darwin-arm64

# Или для Intel Mac
GOOS=darwin GOARCH=amd64 go build -o network-scanner-darwin-amd64
```

## Запуск

После сборки запустите приложение:

```bash
# Для Apple Silicon
./dist/network-scanner-darwin-arm64

# Для Intel
./dist/network-scanner-darwin-amd64

# Или универсальный бинарник (если создан)
./dist/network-scanner-darwin-universal
```

## Разрешения

Для получения MAC адресов может потребоваться запуск с правами администратора:

```bash
sudo ./dist/network-scanner-darwin-arm64
```

## Устранение проблем

### Ошибка: "go: command not found"

Убедитесь, что Go установлен и добавлен в PATH:
```bash
export PATH=$PATH:/usr/local/go/bin
# Или для Homebrew на Apple Silicon:
export PATH=$PATH:/opt/homebrew/bin
```

### Ошибка при сборке зависимостей

Попробуйте очистить кэш и переустановить:
```bash
go clean -modcache
go mod download
```

### Ошибка с правами доступа

Для работы с сетевыми интерфейсами может потребоваться:
- Запуск с `sudo`
- Настройка разрешений в System Preferences > Security & Privacy

