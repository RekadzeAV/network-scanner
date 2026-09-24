# Onboarding

**Дата:** 2026-09-22  

---

## Быстрая настройка (30 минут)

### 1. Клонирование репозитория

```bash
git clone https://github.com/RekadzeAV/network-scanner.git
cd network-scanner
```

### 2. Установка Go

```bash
# Go 1.24+
go version
```

### 3. Установка зависимостей

```bash
go mod download
```

### 4. Сборка

```bash
# CLI
go build -o network-scanner ./cmd/network-scanner

# GUI (требует CGO + Fyne)
go build -o network-scanner-gui ./cmd/gui
```

Или используйте скрипты:
```bash
./scripts/build.sh           # Linux/macOS
./scripts/build.bat          # Windows
./scripts/build-macos.sh     # macOS
```

### 5. Запуск сканирования

```bash
./network-scanner --network 192.168.1.0/24 --ports 80,443
```

### 6. Запуск тестов

```bash
go test ./... -short
```

---

## Структура проекта

| Путь | Описание |
|------|----------|
| `cmd/network-scanner/` | Точка входа CLI |
| `internal/` | Все внутренние пакеты |
| `internal/scanner/` | Ядро сканера |
| `internal/topology/` | Построение топологии |
| `internal/inventory/` | Инвентарь (SQLite) |
| `internal/api/` | REST API |
| `internal/devicecontrol/` | Управление устройствами |
| `internal/remoteexec/` | Удалённое выполнение |
| `internal/security/` | Security service |
| `internal/plugin/` | Плагин-система |
| `internal/eventbus/` | Шина событий |
| `internal/commands/` | Система команд |
| `internal/apperror/` | Единые ошибки |
| `internal/configvalidation/` | Валидация конфигурации |
| `docs/` | Документация |
| `docs/audit/` | Аудиторские артефакты |
| `scripts/` | Скрипты сборки и проверки |

---

## Полезные ссылки

| Ресурс | Путь |
|--------|------|
| Архитектура | `docs/ARCHITECTURE.md` |
| Техническая документация | `docs/TECHNICAL.md` |
| Руководство пользователя | `docs/USER_GUIDE.md` |
| Установка | `docs/INSTALL.md` |
| Roadmap | `docs/ROADMAP.md` |
| Аудит | `docs/audit/` |
| ADR | `docs/adr/` |
| CONTRIBUTING | `CONTRIBUTING.md` |

---

## Отладка

```bash
# С включённым логированием
go build -tags debug -o network-scanner-debug ./cmd/network-scanner

# Профилирование
go test -cpuprofile cpu.prof -memprofile mem.prof ./internal/scanner/...
```

---

## CI/CD

- **CI:** `.github/workflows/ci.yml` — Lint, Test, Build, Closure-checks
- **Release:** `.github/workflows/release.yml` — GoReleaser
- **Pre-commit:** `.pre-commit-config.yaml`