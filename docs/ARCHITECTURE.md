# Архитектура проекта - Network Scanner

**Версия:** 2.0.0  
**Дата обновления:** 2026-01-XX

## Содержание

1. [Обзор архитектуры](#обзор-архитектуры)
2. [Структура проекта](#структура-проекта)
3. [Диаграмма компонентов](#диаграмма-компонентов)
4. [Поток данных](#поток-данных)
5. [Параллелизм и синхронизация](#параллелизм-и-синхронизация)
6. [Обработка ошибок](#обработка-ошибок)
7. [Расширяемость](#расширяемость)
8. [Производительность и оптимизация](#производительность-и-оптимизация)
9. [Безопасность архитектуры](#безопасность-архитектуры)

---

## Обзор архитектуры

Network Scanner использует **модульную архитектуру** с четким разделением ответственности между компонентами и внедрением зависимостей (DI).

### Архитектурные принципы

1. **Разделение ответственности (SoC):**
   - Каждый модуль отвечает за свою область
   - Минимальные зависимости между модулями

2. **Внедрение зависимостей (DI):**
   - Использование `internal/builder` для сборки графа зависимостей
   - Интерфейсы из `internal/contracts` и `internal/scanner/interfaces`

3. **Конкурентность:**
   - Использование горутин Go для параллельной работы
   - Управление ресурсами через семафоры

4. **Расширяемость:**
   - Легко добавлять новые функции
   - Модульная структура позволяет заменять компоненты

5. **Кроссплатформенность:**
   - Абстракция от системных вызовов
   - Использование стандартной библиотеки Go

---

## Структура проекта

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
└── .github/workflows/      # CI/CD
```

---

## Диаграмма компонентов

### CLI приложение

```
┌─────────────────────────────────────────────────────────────┐
│              cmd/network-scanner/main.go                     │
│  - Парсинг параметров командной строки                      │
│  - Инициализация DI контейнера                              │
│  - Обработка сигналов                                        │
│  - Координация работы                                        │
└──────────────┬──────────────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────────────┐
│              internal/builder/                               │
│  - DI Container (NewContainer)                              │
│  - Wiring dependencies                                      │
└──────────────┬──────────────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────────────┐
│              internal/contracts/                             │
│  - Service interfaces (ScannerService, etc.)                │
└──────────────┬──────────────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────────────┐
│              internal/scanner/scanner.go                     │
│  ┌──────────────────────────────────────────────────────┐   │
│  │          NetworkScanner                               │   │
│  │  - Управление процессом сканирования                 │   │
│  │  - Обнаружение активных хостов                        │   │
│  │  - Сканирование портов                                │   │
│  │  - Получение MAC адресов                              │   │
│  │  - Определение типов устройств                        │   │
│  └──────────────────────────────────────────────────────┘   │
└──────────────┬───────────────────────────┬──────────────────┬─────────────────────┘
               │                           │                  │
               ▼                           ▼                  ▼
┌──────────────────────────┐  ┌──────────────────────────────┐  ┌──────────────────────────────┐
│  internal/network/       │  │  internal/presenter/         │  │  internal/snmpcollector/     │
│      network.go          │  │      cli.go                  │  │      collector.go            │
│      prober.go           │  │      json.go                 │  │  - SNMP v2c подключение      │
│      port_scanner.go     │  │      xml.go                  │  │  - ifTable/FDB/LLDP сбор     │
│      parser.go           │  │                              │  │  - CollectWithReport         │
└──────────────────────────┘  └──────────────────────────────┘                 │
                                                                                ▼
                                                                    ┌──────────────────────────────┐
                                                                    │  internal/topology/          │
                                                                    │      topology.go             │
                                                                    │  - BuildTopology             │
                                                                    │  - LLDP/FDB дедуп           │
                                                                    │  - source/confidence links   │
                                                                    │  - JSON/GraphML/DOT export   │
                                                                    └──────────────────────────────┘
```

### GUI приложение

```
┌─────────────────────────────────────────────────────────────┐
│                    cmd/gui/main.go                           │
│  - Точка входа GUI приложения                               │
│  - Инициализация GUI                                        │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│              internal/gui/app.go                             │
│  ┌──────────────────────────────────────────────────────┐   │
│  │                  App (Composition Root)               │   │
│  │  - Wiring вкладок и UI-событий                        │   │
│  │  - Координация подрежимов Devices/Security            │   │
│  │  - Инициализация операций, drawer, analytics          │   │
│  └──────────────────────────────────────────────────────┘   │
└──────────────┬───────────────────────────┬──────────────────┬─────────────────────┘
               │                           │                  │
               ▼                           ▼                  ▼
┌──────────────────────────────┐  ┌──────────────────────────────┐  ┌──────────────────────────────┐
│ internal/gui/controllers     │  │ internal/gui/results         │  │ internal/gui/tools           │
│ - scan_controller.go         │  │ - results_view.go            │  │ - operations.go              │
│ - topology_controller.go     │  │ - results_model.go           │  │ - Operations Center          │
│ - UI state transitions       │  │ - results_analytics_view.go  │  │ - Retry/Cancel              │
└──────────────┬───────────────┘  │ - security_view.go           │  └──────────────┬──────────────┘
               │                  └──────────────┬───────────────┘                 │
               ▼                                 ▼                                 ▼
┌──────────────────────────────┐     ┌──────────────────────────────┐   ┌──────────────────────────────┐
│ internal/scanner/scanner.go  │     │ internal/snmpcollector/      │   │ internal/topology/topology.go│
│ - NetworkScanner             │     │ collector.go                 │   │ - BuildTopology/Export       │
│ - diagnostics summary        │     │ - CollectWithReport          │   │ - Preview/Graph export       │
└──────────────────────────────┘     └──────────────────────────────┘   └──────────────────────────────┘
```

---

## Поток данных

### Основной поток выполнения

```
1. Запуск приложения (main.go)
   │
   ├─► Парсинг параметров командной строки
   │
   ├─► Определение сети (network.go::detectLocalNetwork)
   │   └─► Если не указана, автоматическое определение
   │
   ├─► Создание DI контейнера (builder.NewContainer)
   │   └─► Сборка графа зависимостей (Scanner, Presenter, etc.)
   │
   ├─► Запуск сканирования (scanner.go::Scan)
   │   │
   │   ├─► Парсинг диапазона сети (network.go::parseNetworkRange)
   │   │   └─► Генерация списка IP адресов
   │   │
   │   ├─► Парсинг диапазона портов (network.go::parsePortRange)
   │   │   └─► Генерация списка портов
   │   │
   │   ├─► Обнаружение активных хостов (scanner.go::isHostAlive)
   │   │   ├─► Параллельная проверка каждого IP
   │   │   └─► Проверка популярных портов (80, 443, 22, ...)
   │   │
   │   └─► Сканирование портов на активных хостах
   │       ├─► Для каждого активного хоста (scanner.go::scanHost)
   │       │   ├─► Получение MAC адреса (scanner.go::getMACAddress)
   │       │   │   └─► ARP запрос через pcap (scanner.go::getMACViaARPRequest)
   │       │   │
   │       │   ├─► Получение hostname (net.LookupAddr)
   │       │   │
   │       │   ├─► Сканирование портов (network.go::isPortOpen)
   │       │   │   └─► TCP connect для каждого порта
   │       │   │
   │       │   ├─► Определение протоколов (scanner.go::getProtocolFromPort)
   │       │   │
   │       │   └─► Определение типа устройства (scanner.go::detectDeviceType)
   │       │
   │       └─► Сохранение результатов в ScanResult
   │
   ├─► Получение результатов (scanner.go::GetResults)
   │
   ├─► Отображение результатов (presenter::CLIPresenter)
   │   └─► Форматирование в таблицу
   │
   ├─► Отображение аналитики (display.go::displayAnalytics)
   │   ├─► Статистика по протоколам
   │   ├─► Статистика по портам
   │   ├─► Статистика по типам устройств
   │   └─► Общая статистика
   │
   └─► [опционально] режим топологии (--topology)
       ├─► SNMP сбор (snmpcollector.go::CollectWithReport)
       │   ├─► Подключение к SNMP target
       │   ├─► sysName/sysDescr
       │   ├─► ifTable/FDB/LLDP
       │   └─► Формирование CollectReport (connected/partial/failed)
       ├─► Построение графа (topology.go::BuildTopology)
       │   ├─► Обогащение устройств SNMP-данными
       │   ├─► LLDP/FDB связи
       │   ├─► Дедуп и confidence
       │   └─► Фильтрация broadcast/multicast/zero MAC
       └─► Экспорт/вывод
           ├─► stdout (кратко)
           ├─► JSON / GraphML
           └─► PNG / SVG через Graphviz dot
```

### Поток данных при сканировании хоста

```
IP адрес
   │
   ▼
┌─────────────────────┐
│  isHostAlive()      │
│  Проверка портов:   │
│  80, 443, 22, ...   │
└──────────┬──────────┘
           │
           ├─► Хост не отвечает ──► Пропуск
           │
           └─► Хост отвечает ──► Продолжение
                   │
                   ▼
           ┌─────────────────────┐
           │  scanHost()         │
           └──────────┬──────────┘
                      │
        ┌─────────────┼─────────────┐
        │             │             │
        ▼             ▼             ▼
   ┌─────────┐  ┌──────────┐  ┌──────────┐
   │  MAC    │  │ Hostname │  │  Порты   │
   │  адрес  │  │          │  │          │
   └────┬────┘  └────┬─────┘  └────┬─────┘
        │            │             │
        └────────────┼─────────────┘
                     │
                     ▼
            ┌─────────────────┐
            │  ScanResult     │
            │  - IP           │
            │  - MAC          │
            │  - Hostname     │
            │  - Ports[]      │
            │  - Protocols[]  │
            │  - DeviceType   │
            │  - Vendor       │
            └─────────────────┘
```

### Поток данных в режиме топологии

```
scanner.Result[]
   │
   ▼
snmpcollector.CollectWithReport()
   │
   ├─► map[string]*topology.Device   (SNMP данные)
   └─► CollectReport                 (quality/ошибки SNMP)
               │
               ▼
topology.BuildTopology(results, snmpData)
   │
   └─► topology.Topology
          ├─► Devices
          └─► Links (SourceType + Confidence + Evidence)
```

---

## Параллелизм и синхронизация

### Модель параллелизма

Проект использует **worker pool pattern** с горутинами Go:

```
┌─────────────────────────────────────────┐
│         Main Goroutine                  │
│  (Управление и координация)             │
└──────────────┬──────────────────────────┘
               │
               ▼
    ┌──────────────────────┐
    │   Semaphore Channel   │
    │   (Ограничение        │
    │    параллелизма)      │
    └──────────┬────────────┘
               │
    ┌──────────┼──────────┐
    │          │          │
    ▼          ▼          ▼
┌────────┐ ┌────────┐ ┌────────┐
│Worker 1│ │Worker 2│ │Worker N│
│Goroutine│ │Goroutine│ │Goroutine│
└────────┘ └────────┘ └────────┘
    │          │          │
    └──────────┼──────────┘
               │
               ▼
    ┌──────────────────────┐
    │   Results Collection  │
    │   (С синхронизацией)  │
    └──────────────────────┘
```

### Синхронизация

1. **Семафор для ограничения параллелизма:**
   ```go
   sem := make(chan struct{}, ns.threads)
   ```

2. **Мьютексы для защиты общих данных:**
   ```go
   var mu sync.RWMutex  // Для результатов
   var aliveMutex sync.Mutex  // Для списка активных хостов
   ```

3. **WaitGroup для ожидания завершения:**
   ```go
   var wg sync.WaitGroup
   ```

4. **Context для отмены операций:**
   ```go
   ctx, cancel := context.WithCancel(context.Background())
   ```

### Пример параллельного сканирования

```go
// Создание пула горутин
sem := make(chan struct{}, threads)

// Для каждого IP
for _, ip := range ips {
    sem <- struct{}{}  // Захват семафора
    wg.Add(1)
    
    go func(ip net.IP) {
        defer func() { <-sem }()  // Освобождение семафора
        defer wg.Done()
        
        // Работа с IP
        if isHostAlive(ip) {
            mu.Lock()
            aliveIPs = append(aliveIPs, ip)
            mu.Unlock()
        }
    }(ip)
}

wg.Wait()  // Ожидание завершения всех горутин
```

---

## Обработка ошибок

### Стратегия обработки ошибок

1. **Проверка на каждом этапе:**
   - Парсинг параметров
   - Определение сети
   - Подключения к хостам

2. **Graceful degradation:**
   - Если MAC адрес не получен - продолжаем без него
   - Если hostname не получен - используем IP
   - Если устройство не отвечает - пропускаем

3. **SNMP partial failure policy:**
   - `connect_error` -> устройство помечается failed
   - `query_error` -> устройство учитывается как partial
   - Топология строится из доступных данных, без fail-fast на одном OID

4. **Логирование ошибок:**
   - Вывод в консоль для важных ошибок
   - Тихий пропуск для некритичных ошибок

### Примеры обработки ошибок

```go
// Определение сети
network, err := detectLocalNetwork()
if err != nil {
    log.Fatalf("Не удалось определить сеть: %v", err)
}

// Получение MAC (не критично)
mac, err := ns.getMACAddress(ip)
if err == nil {
    result.MAC = mac
}
// Продолжаем даже если MAC не получен

// Проверка порта
if isPortOpen(host, port, timeout) {
    // Порт открыт
} else {
    // Порт закрыт или недоступен - пропускаем
}
```

---

## Расширяемость

### Точки расширения

1. **Добавление новых методов обнаружения хостов:**
   ```go
   // В scanner.go можно добавить:
   func (ns *NetworkScanner) isHostAliveICMP(ip string) bool
   func (ns *NetworkScanner) isHostAliveARP(ip string) bool
   ```

2. **Расширение определения устройств:**
   ```go
   // В scanner.go можно улучшить:
   func (ns *NetworkScanner) detectDeviceType(result ScanResult) string {
       // Добавить больше правил
   }
   ```

3. **Добавление новых форматов экспорта:**
   ```go
   // Новый файл export.go:
   func exportToJSON(results []ScanResult) error
   func exportToCSV(results []ScanResult) error
   ```

4. **Расширение SNMP источников топологии:**
   ```go
   // В snmpcollector можно добавить:
   func GetCDPNeighbors() (..., error)
   func GetVendorMIBNeighbors() (..., error)
   ```

### Интерфейсы для расширения

Можно создать интерфейсы для плагинов:

```go
type PortScanner interface {
    ScanPort(host string, port int) PortInfo
}

type DeviceDetector interface {
    DetectDevice(result ScanResult) DeviceInfo
}

type ResultExporter interface {
    Export(results []ScanResult) error
}
```

---

## Производительность и оптимизация

### Оптимизации

1. **Двухэтапное сканирование:**
   - Сначала быстрая проверка доступности
   - Затем детальное сканирование только активных

2. **Параллелизм:**
   - Настраиваемое количество потоков
   - Эффективное использование ресурсов

3. **Минимизация сетевых запросов:**
   - Проверка только необходимых портов
   - Настраиваемый таймаут
   - SNMP опрос только для устройств с открытым UDP/161

### Метрики производительности

- **Время проверки хоста:** ~100-300ms
- **Время сканирования порта:** ~50-200ms
- **Пропускная способность:** ~100-500 хостов/минуту (зависит от настроек)

---

## Безопасность архитектуры

### Защита от race conditions

- Использование мьютексов для общих данных
- Атомарные операции где возможно
- Правильное использование каналов

### Защита ресурсов

- Ограничение параллелизма через семафоры
- Таймауты для всех сетевых операций
- Graceful shutdown при прерывании

### Надежность GUI при долгих операциях

- Кнопки построения/сохранения топологии блокируются на время SNMP+Build фазы
- Это предотвращает повторный запуск и race-condition в UI
- Tool-операции выполняются через `Operations Runtime`:
  - единые статусы (`queued/running/success/failed/canceled`),
  - централизованный `Retry/Cancel` в `Operations Center`
  - снижает риск несогласованных состояний при параллельных действиях пользователя

---

## Дополнительные ресурсы

- [Инструкция по эксплуатации](../Инструкция%20по%20эксплуатации.md) - Полная инструкция по эксплуатации (русский язык)
- [Техническая документация](TECHNICAL.md) - Техническая документация для разработчиков
- [Руководство пользователя](USER_GUIDE.md) - Подробное руководство пользователя
- [GUI документация](GUI.md) - Документация по GUI версии
- [Инструкция по установке](INSTALL.md) - Инструкции по установке
- [README.md](../README.md) - Основная документация проекта
- [BUILD_STRUCTURE.md](BUILD_STRUCTURE.md) - Структура каталогов релизной сборки (`build/release/`)
- [RELEASE_OPERATIONS_CHEATSHEET.md](RELEASE_OPERATIONS_CHEATSHEET.md) - Команды closure и локальные релизные артефакты
- [ROADMAP.md](ROADMAP.md) - Дорожная карта проекта
- [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) - План реализации

---

**Версия документа:** 2.0.0  
**Последнее обновление:** 2026-01-XX

