# Структура проекта Network Scanner

**Версия:** 2.0.0  
**Дата обновления:** 2026-01-XX

## Текущая структура (актуальная)

```
network-scanner/
├── cmd/
│   ├── network-scanner/    # CLI точка входа
│   │   ├── main.go         # Точка входа CLI
│   │   ├── cmd/            # Команды CLI (scan, security, topology, и т.д.)
│   │   └── dpi_windows.go  # DPI awareness для Windows
│   └── gui/                # GUI точка входа
├── internal/
│   ├── builder/            # DI Container и сборка зависимостей
│   ├── contracts/          # Service interfaces (общие интерфейсы)
│   ├── services/           # Service wrappers и factory
│   ├── scanner/            # Ядро сканирования
│   │   ├── scanner.go      # NetworkScanner
│   │   ├── interfaces.go   # Интерфейсы (NetworkProber, PortScanner, и т.д.)
│   │   ├── deviceclassifier/ # Определение типов устройств
│   │   └── daemon/         # Daemon mode для фоновой работы
│   ├── network/            # Сетевые операции
│   │   ├── network.go      # Определение сети, парсинг CIDR
│   │   ├── prober.go       # NetworkProber (ping, ARP)
│   │   ├── port_scanner.go # PortScanner (TCP/UDP)
│   │   └── parser.go       # Парсинг диапазонов
│   ├── display/            # Вывод результатов (CLI)
│   ├── presenter/          # Презентеры (CLI, JSON, XML)
│   ├── gui/                # GUI компоненты (Fyne)
│   │   ├── app.go          # Главный файл GUI
│   │   ├── scan_controller.go    # Контроллер сканирования
│   │   ├── topology_controller.go # Контроллер топологии
│   │   ├── results_view.go       # Отображение результатов
│   │   └── operations.go         # Operations Runtime
│   ├── topology/           # Построение топологии сети
│   ├── snmpcollector/      # SNMP сбор данных
│   ├── security/           # Анализ безопасности
│   ├── audit/              # Аудит открытых портов
│   ├── risksignature/      # Risk Signatures
│   ├── devicecontrol/      # Device Control (HTTP API)
│   ├── remoteexec/         # Remote Exec (SSH/WMI/WinRM)
│   ├── inventory/          # Инвентаризация (SQLite)
│   ├── comparator/         # Сравнение снапшотов
│   ├── alerting/           # Система уведомлений
│   ├── api/                # REST API
│   ├── report/             # Экспорт отчетов (PDF/HTML)
│   ├── banner/             # Banner grabbing
│   ├── osdetect/           # Определение ОС
│   ├── ports/              # Базы портов и сервисов
│   ├── cache/              # Кэширование (DNS, MAC)
│   ├── batch/              # Батч-обработка (SNMP)
│   ├── profiler/           # CPU/Memory profiling
│   ├── wol/                # Wake-on-LAN
│   ├── nettools/           # Сетевые инструменты (ping, traceroute, dns, whois, wifi)
│   ├── diff/               # Сравнение сканирований
│   ├── redact/             # Маскирование чувствительных данных
│   ├── errors/             # Единая система ошибок
│   ├── logger/             # Логирование
│   ├── mock/               # Mock-сервисы для тестирования
│   └── legacy/             # Архивированный код
├── scripts/                # Скрипты сборки и тестов
├── docs/                   # Документация
├── config/                 # Конфигурационные файлы
├── inventory/              # База данных инвентаризации (SQLite)
├── build/                  # Локальные артефакты (в .gitignore)
│   └── release/            # Выход релизных скриптов (YYYY-MM-DD-N/)
├── .github/workflows/      # CI/CD
├── go.mod                  # Зависимости проекта
├── go.sum                  # Checksums зависимостей
├── .gitignore              # Игнорируемые файлы
└── README.md               # Корневой README
```

## Преимущества структуры

1. **Соответствие стандартам Go:**
   - Стандартная структура Go проектов
   - `cmd/` для точек входа
   - `internal/` для внутренних пакетов

2. **Модульность и разделение ответственности:**
   - Каждый модуль отвечает за свою область
   - DI Container для сборки зависимостей
   - Интерфейсы из `internal/contracts/`

3. **Расширяемость:**
   - Легко добавлять новые модули
   - Четкое разделение кода и документации
   - Удобнее для поддержки

4. **Тестируемость:**
   - Mock-сервисы в `internal/mock/`
   - Unit-тесты для каждого пакета
   - Integration-тесты с build tag

## Связанные документы

- [ARCHITECTURE.md](ARCHITECTURE.md) - Архитектура проекта (v2.0)
- [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) - План реализации v2.0
- [USER_GUIDE.md](USER_GUIDE.md) - Руководство пользователя
- [TECHNICAL.md](TECHNICAL.md) - Техническая документация

---

**Версия документа:** 2.0.0  
**Последнее обновление:** 2026-01-XX

