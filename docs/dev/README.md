# Developer documentation (`docs/dev/`)

Каталог для сгенерированной и справочной документации для разработчиков.

## Автогенерация godoc

Скрипт `scripts/gen-godoc.ps1` (Windows) / `scripts/gen-godoc.sh` (Linux/macOS)
генерирует API-документацию по пакетам `internal/*` и `cmd/*`.

```powershell
# Windows
pwsh -ExecutionPolicy Bypass -File .\scripts\gen-godoc.ps1
```

```bash
# Linux / macOS
./scripts/gen-godoc.sh
```

Поведение:

1. Если установлен `godoc` (`go install golang.org/x/tools/cmd/godoc@latest`),
   генерируется HTML-индекс в `docs/dev/godoc/`.
2. Иначе создаётся текстовый дамп `docs/dev/godoc/GODOC_DUMP.txt` через `go doc -all`
   по всем пакетам `internal/*` и `cmd/*`.

## Сгенерированные страницы (v2.4)

Ниже — список пакетов, для которых сгенерирована справка в `docs/dev/godoc/GODOC_DUMP.txt`:

### `internal/`

| Пакет | Описание |
|-------|----------|
| `alerting` | Движок алертов (Engine, Alert, ConsoleHandler) |
| `api` | REST API (gorilla/mux, Bearer-токен, CORS, rate-limit) |
| `apperror` | Типизированные ошибки приложения |
| `audit` | Аудит открытых портов |
| `banner` | Banner grabbing |
| `batch` | Пакетная обработка SNMP |
| `benchmark` | Бенчмарки и perf-regression gate |
| `builder` | DI-контейнер и сборка зависимостей |
| `cache` | Кэширование (DNS, MAC) |
| `commands` | Командный слой (Command/Request/Response) |
| `comparator` | Сравнение снапшотов |
| `configvalidation` | Схемы и валидаторы конфигурации |
| `contracts` | Service interfaces (общие интерфейсы) |
| `cve` | Поиск уязвимостей (NVD-базы) |
| `devicecontrol` | Управление устройствами (HTTP API) |
| `diff` | Сравнение сканирований |
| `display` | Вывод результатов (CLI) |
| `errors` | Единая система ошибок |
| `eventbus` | Шина событий (E6) |
| `features` | Feature flags (D1–D3) |
| `gui` | GUI-слой (Fyne) |
| `gui/controller` | Контроллеры GUI (scan, security, results, topology, tools) |
| `gui/errors` | Ошибки GUI-слоя |
| `gui/inventory` | Представление инвентаря |
| `inventory` | Инвентаризация (SQLite) |
| `logger` | Логирование |
| `metrics` | Реестр метрик (Counter/Gauge/Histogram), Prometheus exposition, /metrics endpoint |
| `mock` | Mock-сервисы для тестирования |
| `nettools` | Сетевые инструменты (ping, traceroute, dns, whois, wifi) |
| `network` | Сетевые операции (CIDR, prober, port_scanner) |
| `osdetect` | Определение ОС |
| `plugin` | Система плагинов (Loader, Manifest, Plugin interface) |
| `ports` | Базы портов и сервисов |
| `presenter` | Презентеры (CLI, JSON, XML) |
| `profiler` | CPU/Memory profiling |
| `redact` | Маскирование чувствительных данных |
| `remoteexec` | Remote Exec (SSH/WMI/WinRM) |
| `report` | Экспорт отчётов (PDF/HTML) |
| `risksignature` | Risk Signatures |
| `scanner` | Ядро сканирования |
| `scanner/daemon` | Daemon mode для фоновой работы |
| `scanner/deviceclassifier` | Определение типов устройств |
| `scanner/plugin` | Плагин-сканер |
| `security` | Анализ безопасности |
| `services` | Service wrappers и factory |
| `snmpcollector` | SNMP сбор данных |
| `telemetry` | Телеметрия (метрики, opt-in) |
| `topology` | Построение топологии сети |
| `wol` | Wake-on-LAN |

### `cmd/`

| Пакет | Описание |
|-------|----------|
| `gui` | GUI точка входа |
| `md2pdf` | Конвертация markdown в PDF |
| `network-scanner` | CLI точка входа + подкоманды cobra |

## Где искать документацию

| Документ | Назначение |
|----------|------------|
| [../ARCHITECTURE.md](../ARCHITECTURE.md) | Архитектура, DI-контейнер, событийная шина, метрики, плагины |
| [../TECHNICAL.md](../TECHNICAL.md) | Техническая документация для разработчиков |
| [../PROJECT_STRUCTURE.md](../PROJECT_STRUCTURE.md) | Структура каталогов |
| [../swagger.yaml](../swagger.yaml) | Спецификация REST API (включая Bearer-аутентификацию) |
| [../ROADMAP.md](../ROADMAP.md) | Дорожная карта проекта |
| [CONTRIBUTING.md](../../CONTRIBUTING.md) | Процесс контрибьюции, code review |

> **Примечание:** содержимое `docs/dev/godoc/` — генерируемый артефакт и не
> коммитится в репозиторий (см. `.gitignore`).
