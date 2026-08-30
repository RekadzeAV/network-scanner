# Инструкция по установке и сборке — Network Scanner v2.1

**Версия:** 2.1.0  
**Дата обновления:** 2026-08-29

## Системные требования

- **Go 1.23+** (рекомендуется последняя стабильная версия)
- **ОС:** Windows 10+, macOS 12+, Linux (Ubuntu 20+, Debian 11+)
- **Для GUI:** X11/Wayland (Linux), Native display (macOS/Windows)

---

## Установка Go

### Windows

1. Скачайте установщик с https://go.dev/dl/
2. Запустите установщик (`go1.xx.x.windows-amd64.msi`)
3. Добавьте `C:\Program Files\Go\bin` в PATH (если не добавилось автоматически)
4. Проверьте: `go version`

### macOS

```bash
# Через Homebrew (рекомендуется)
brew install go

# Или прямая установка
# Перейдите на https://go.dev/dl/, скачайте установщик для macOS
```

Проверьте: `go version`

### Linux (Ubuntu/Debian)

```bash
# Скачайте последнюю версию Go
wget https://go.dev/dl/go1.23.linux-amd64.tar.gz

# Распакуйте в /usr/local
sudo tar -C /usr/local -xzf go1.23.linux-amd64.tar.gz

# Добавьте в PATH (добавьте в ~/.bashrc или ~/.zshrc)
export PATH=$PATH:/usr/local/go/bin

# Примените изменения
source ~/.bashrc
```

Проверьте: `go version`

---

## Сборка приложения

### Быстрая сборка

```bash
# Перейдите в директорию проекта
cd network-scanner

# Установка зависимостей
go mod download

# Сборка CLI
go build -o network-scanner ./cmd/network-scanner

# Сборка GUI
go build -o network-scanner-gui ./cmd/gui
```

### Кроссплатформенная сборка

```bash
# Windows (из Linux/macOS)
GOOS=windows GOARCH=amd64 go build -o network-scanner.exe ./cmd/network-scanner

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o network-scanner-darwin-arm64 ./cmd/network-scanner

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o network-scanner-darwin-amd64 ./cmd/network-scanner

# Linux
GOOS=linux GOARCH=amd64 go build -o network-scanner-linux-amd64 ./cmd/network-scanner
```

---

## Запуск

### CLI

```bash
# Windows
.\network-scanner.exe scan --network 192.168.1.0/24

# macOS/Linux
./network-scanner scan --network 192.168.1.0/24
```

### GUI

```bash
# Windows
.\network-scanner-gui.exe

# macOS/Linux
./network-scanner-gui
```

---

## Разрешения

### Windows

Для полного сканирования (ARP, MAC-адреса) запустите от имени администратора:
- Правый клик на ярлык → "Запуск от имени администратора"
- Или через PowerShell: `Start-Process network-scanner.exe -Verb RunAs`

### macOS

```bash
# Запуск с sudo
sudo ./network-scanner scan --network 192.168.1.0/24
```

### Linux

```bash
# Вариант 1: Запуск с sudo
sudo ./network-scanner scan --network 192.168.1.0/24

# Вариант 2: Установка capability (рекомендуется)
sudo setcap cap_net_raw+ep ./network-scanner
```

---

## Тестирование

```bash
# Запуск всех тестов
go test ./...

# Запуск тестов с race detector
go test -race ./...

# Запуск тестов для конкретного пакета
go test ./internal/scanner/...

# Покрытие тестов
go test -cover ./...
```

---

## Устранение проблем

### Ошибка: "go: command not found"

Убедитесь, что Go установлен и добавлен в PATH:

```bash
# Windows
$env:PATH += ";C:\Program Files\Go\bin"

# macOS/Linux
export PATH=$PATH:/usr/local/go/bin
# Или для Homebrew на Apple Silicon:
export PATH=$PATH:/opt/homebrew/bin
```

### Ошибка при сборке зависимостей

```bash
go clean -modcache
go mod download
```

### Ошибка с правами доступа

- Запустите с `sudo` (Linux/macOS) или от имени администратора (Windows)
- Для Linux: `sudo setcap cap_net_raw+ep ./network-scanner`

### GUI не запускается (Linux)

Убедитесь, что установлены зависимости Fyne:

```bash
# Ubuntu/Debian
sudo apt-get install libgl1-mesa-dev libxcb-xinerama0

# Fedora
sudo dnf install mesa-libGL libxcb-devel
```

---

## Документация

- **Архитектура:** [docs/ARCHITECTURE.md](ARCHITECTURE.md)
- **Дорожная карта:** [ROADMAP.md](../ROADMAP.md)
- **Список задач:** [docs/TASK_BACKLOG_V21.md](TASK_BACKLOG_V21.md)
- **Чеклист GUI:** [docs/GUI_SMOKE_CHECKLIST.md](GUI_SMOKE_CHECKLIST.md)

---

**Версия документа:** 2.1.0  
**Последнее обновление:** 2026-08-29
