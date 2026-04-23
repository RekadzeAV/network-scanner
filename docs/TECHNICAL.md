# Техническая документация - Network Scanner

## Содержание

1. [Обзор архитектуры](#обзор-архитектуры)
2. [Структура проекта](#структура-проекта)
3. [Основные компоненты](#основные-компоненты)
4. [Алгоритмы и методы](#алгоритмы-и-методы)
5. [Зависимости](#зависимости)
6. [Производительность](#производительность)
7. [Ограничения](#ограничения)
8. [Будущие улучшения](#будущие-улучшения)

---

## Обзор архитектуры

Network Scanner построен на языке Go и использует конкурентную модель с горутинами для параллельного сканирования сети.

### Основные принципы

- **Многопоточность:** Использует горутины Go для параллельного сканирования
- **Модульность:** Разделение на логические модули (сканирование, сеть, отображение)
- **Кроссплатформенность:** Работает на Windows, macOS, Linux
- **Расширяемость:** Легко добавлять новые функции

### Технологический стек

- **Язык:** Go 1.24+
- **Библиотеки:**
  - `github.com/google/gopacket` - работа с сетевыми пакетами (ARP)
  - `github.com/jedib0t/go-pretty/v6` - форматирование таблиц (CLI)
  - `github.com/gosnmp/gosnmp` - SNMP v2c опрос устройств
  - `fyne.io/fyne/v2` - кроссплатформенный GUI фреймворк
  - Стандартная библиотека Go для сетевых операций

---

## Структура проекта

```
Сканер локальной сети/
├── cmd/
│   ├── network-scanner/
│   │   └── main.go          # Точка входа CLI приложения
│   └── gui/
│       └── main.go          # Точка входа GUI приложения
├── internal/
│   ├── scanner/
│   │   └── scanner.go       # Основная логика сканирования
│   ├── network/
│   │   └── network.go       # Работа с сетью (определение сети, парсинг)
│   ├── display/
│   │   └── display.go       # Отображение результатов и аналитика (CLI)
│   └── gui/
│       ├── app.go                     # Composition root GUI и wiring вкладок
│       ├── scan_controller.go         # UI-state сканирования
│       ├── topology_controller.go     # UI-state топологии
│       ├── operations.go              # Runtime операций (Run/Cancel/Retry)
│       ├── results_view.go            # Devices/Security подрежимы, filters, drawer
│       ├── results_model.go           # Базовый пайплайн фильтрации/сортировки
│       ├── results_analytics_view.go  # Явная аналитика в пайплайне рендера
│       ├── results_charts.go          # Pie charts + cache
│       ├── security_view.go           # Security Dashboard + HTML export
│       └── formatter.go               # Доп. форматирование GUI представлений
├── docs/                    # Документация
├── scripts/                 # Скрипты сборки
├── go.mod                   # Зависимости проекта
└── README.md               # Основная документация
```

---

## Основные компоненты

### 1. cmd/network-scanner/main.go - Точка входа CLI

**Ответственность:**
- Парсинг параметров командной строки
- Инициализация сканера
- Обработка сигналов (Ctrl+C)
- Координация работы компонентов

**Ключевые функции:**
- `main()` - главная функция CLI приложения

**Параметры командной строки:**
```go
- network: string          // Диапазон сети (CIDR)
- timeout: int             // Таймаут TCP/UDP в секундах
- ports: string            // Диапазон портов
- threads: int             // Количество потоков
- show-closed: bool        // Показывать закрытые порты
- udp: bool                // Включить UDP-сканирование
- topology: bool           // Построение топологии
- output-format: string    // graphml|png|svg|json
- output-file: string      // Путь к файлу экспорта
- snmp-community: string   // Список community через запятую
- snmp-timeout: int        // Таймаут SNMP в секундах
- ping: string             // Ping выбранного хоста/IP и выход
- traceroute: string       // Traceroute/tracert выбранного хоста/IP и выход
- dns: string              // DNS lookup выбранного имени/IP и выход
- whois: string            // Whois lookup выбранного имени/IP и выход
- wifi: bool               // Показать Wi-Fi информацию текущей ОС и выход
- dns-server: string       // Кастомный resolver для --dns
- ping-count: int          // Количество ping-пакетов
- tool-timeout: int        // Таймаут tool-режимов в секундах
- traceroute-max-hops: int // Максимальное число hops для traceroute
- raw: bool                // Печатать raw output инструментов
- grab-banners: bool       // Собирать баннеры/версии сервисов
- show-raw-banners: bool   // Печатать сырой banner в CLI выводе
- wol-mac: string          // Отправить Wake-on-LAN magic packet и выход
- wol-broadcast: string    // Broadcast для --wol-mac
- wol-iface: string        // Интерфейс для автоподбора broadcast в WOL
- os-detect-active: bool   // Включить расширенные (active) эвристики определения ОС
- audit-open-ports: bool   // Базовый аудит открытых портов после сканирования
- audit-min-severity: string // Минимальная критичность для аудита: all|low|medium|high|critical
- risk-signatures: bool    // Локальные сигнатуры домашних рисков
- device-action: string    // Управление устройством: status|reboot
- device-target: string    // URL API устройства для device-control
- device-vendor: string    // Профиль API: generic-http|tp-link-http
- device-user: string      // Username для device-control
- device-pass: string      // Password для device-control
- device-confirm: string   // Подтверждение reboot: I_UNDERSTAND
- device-timeout: int      // Таймаут device-control в секундах
- audit-log: string        // Путь JSONL audit log для device-control
- cve: bool                // Базовое CVE сопоставление
- cve-min-cvss: float64    // Минимальный CVSS фильтр
- cve-max-age-days: int    // Максимальный возраст CVE в днях
- security-report-file: string   // HTML security report
- security-report-redact: bool   // Маскирование чувствительных данных в report
- remote-exec-*: ...       // Набор флагов remote execution (SSH/WMI/WinRM)
```

### 2. cmd/gui/main.go - Точка входа GUI

**Ответственность:**
- Инициализация GUI приложения
- Запуск GUI интерфейса

**Ключевые функции:**
- `main()` - главная функция GUI приложения

### 3. internal/gui/* - GUI архитектура

**Ответственность (по модулям):**
- `app.go`:
  - инициализация окна и вкладок (`Сканирование`, `Топология`, `Инструменты`);
  - wiring UI-компонентов и маршрутизация событий.
- `scan_controller.go`:
  - переходы scan UI-state (`start/completion/timeout`) и синхронизация статусов.
- `topology_controller.go`:
  - переходы topology UI-state (`progress/canceled/failure/success`).
- `operations.go`:
  - runtime операций для tools/долгих задач (`queued/running/success/failed/canceled`);
  - действия `Run`, `Cancel`, `Retry`, подписка на обновления.
- `results_view.go` + `results_model.go`:
  - базовый пайплайн фильтрации/сортировки;
  - рендер `Devices` (`Таблица`/`Карточки`) и `Host Details Drawer`.
- `security_view.go`:
  - `Security` подрежим с агрегированными findings (`audit + risk signatures`) и HTML export.
- `results_analytics_view.go` + `results_charts.go`:
  - встроенная аналитика в пайплайне рендера (`markdown summary` / `pie charts`).

**Ключевые типы (сокращенно):**

```go
type App struct {
    myApp              fyne.App
    myWindow           fyne.Window
    scanResults        []scanner.Result
    networkScanner     *scanner.NetworkScanner
    operations         *OperationsManager
    resultsSubMode     string // Devices|Security
    selectedHostIP     string // Host Details Drawer
}
```

**Ключевые методы/потоки:**
- `NewApp()` / `initUI()` / `setupEventHandlers()`
- `startScan()` + scan-state методы в `scan_controller.go`
- `buildTopology()` + topology-state методы в `topology_controller.go`
- tool-операции через `runToolOperation(...)` и `OperationsManager`
- `renderScanResultsView()` с подрежимами `Devices/Security`
- `saveResults()` / `saveTopology()` / `savePerformanceReport()`

**Responsive layout policy (оконный/fullscreen):**
- Профиль интерфейса вычисляется по ширине canvas:
  - `compact` (`<= 1366px`)
  - `normal` (`1367..2199px`)
  - `wide` (`>= 2200px`)
- Профиль применяется через единый runtime-хук `applyResponsiveLayout(...)`.
- Вкладки `Сканирование`, `Топология`, `Инструменты` используют вертикальные scroll-контейнеры для панелей управления, чтобы контролы не "терялись" на малой высоте окна.
- В `compact` режиме:
  - `Host Details` рендерится в вертикальном split;
  - таблица результатов использует укороченные заголовки и узкие колонки;
  - grid-блоки фильтров/сортировки/operations перестраиваются в 1-2 колонки.
- В `wide` режиме расширяются размеры рабочих областей и ширины таблиц.

**Инварианты UI-архитектуры:**
- Функционал должен оставаться одинаково доступным в оконном и полноэкранном режимах.
- Новые UI-блоки должны иметь fallback для `compact` (stack/scroll/уменьшение числа колонок).
- Не добавлять "жесткие" размеры без проверки на матрице разрешений (`1366x768` и выше, включая high-DPI).

### 4. internal/gui/formatter.go - Форматирование для GUI

**Ответственность:**
- Форматирование результатов сканирования для отображения в GUI
- Генерация Markdown разметки для результатов

**Ключевые функции:**
- `FormatResultsForDisplay()` - форматирует результаты в Markdown формат

### 5. internal/scanner/scanner.go - Логика сканирования

**Ответственность:**
- Управление процессом сканирования
- Обнаружение активных хостов
- Сканирование портов
- Определение MAC адресов
- Определение типов устройств

**Основные типы:**

```go
// Результат сканирования одного хоста
type ScanResult struct {
    IP          string
    MAC         string
    Hostname    string
    Ports       []PortInfo
    Protocols   []string
    DeviceType  string
    DeviceVendor string
    IsAlive     bool
    GuessOS     string
    GuessOSConfidence string
    GuessOSReason string
}

// Информация о порте
type PortInfo struct {
    Port     int
    State    string  // "open", "closed", "filtered"
    Protocol string  // "tcp", "udp"
    Service  string
    Version  string  // нормализованная версия/сигнатура
    Banner   string  // сырой banner (опционально)
}

// Сканер сети
type NetworkScanner struct {
    network   string
    timeout   time.Duration
    portRange string
    threads   int
    results   []ScanResult
    mu        sync.RWMutex
    ctx       context.Context
    cancel    context.CancelFunc
    wg        sync.WaitGroup
}
```

### 8. internal/snmpcollector/collector.go - SNMP сборщик

**Ответственность:**
- Подключение к устройствам по SNMP v2c
- Сбор `sysName`, `sysDescr`, `ifTable`, FDB (MAC-таблица), LLDP соседей
- Формирование структуры данных для построения топологии
- Формирование отчета по частичным/полным SNMP отказам

**Ключевые типы/методы:**
- `SNMPClient`, `GoSNMPClient`
- `Collect(...)` - совместимый API
- `CollectWithReport(...)` - расширенный API с `CollectReport`
- `CollectReport`:
  - `TotalSNMPTargets`, `Connected`, `Partial`, `Failed`, `Failures[]`
- `DeviceFailure`:
  - `Kind`: `connect_error` или `query_error`

### 9. internal/topology/topology.go - Построение топологии и экспорт

**Ответственность:**
- Обогащение результатов сканирования SNMP-данными
- Построение связей из LLDP/FDB с дедупликацией
- Нормализация MAC и фильтрация broadcast/multicast/zero MAC
- Экспорт в JSON, GraphML, DOT, PNG/SVG (через Graphviz)

**Ключевые типы/методы:**
- `BuildTopology(...)`
- `Topology.SaveJSON(...)`
- `Topology.SaveGraphML(...)`
- `Topology.ToDOT(...)`
- `Topology.RenderWithGraphviz(...)`
- `Topology.Validate(...)` — schema-check перед экспортом
- `Link.SourceType`: `lldp|fdb|inferred`
- `Link.Confidence`: `high|medium|low`

**Ключевые методы:**

- `NewNetworkScanner()` - создание нового сканера
- `Scan()` - запуск сканирования
- `isHostAlive()` - проверка доступности хоста
- `scanHost()` - сканирование одного хоста
- `getMACAddress()` - получение MAC адреса
- `detectDeviceType()` - определение типа устройства
- `Stop()` - остановка сканирования
- `GetResults()` - получение результатов

### 6. internal/network/network.go - Работа с сетью

**Ответственность:**
- Автоматическое определение локальной сети
- Парсинг диапазонов сети (CIDR)
- Парсинг диапазонов портов
- Проверка открытости портов
- Определение сервисов по портам

**Ключевые функции:**

- `detectLocalNetwork()` - автоматическое определение сети
- `parseNetworkRange()` - парсинг CIDR нотации
- `parsePortRange()` - парсинг диапазона портов
- `isPortOpen()` - проверка открытости порта
- `getServiceName()` - определение сервиса по порту

### 7. internal/display/display.go - Отображение результатов (CLI)

**Ответственность:**
- Форматирование и вывод результатов
- Генерация аналитики
- Статистика по протоколам и портам

**Ключевые функции:**

- `displayResults()` - вывод таблицы результатов
- `displayAnalytics()` - вывод аналитики
- `formatPorts()` - форматирование списка портов
- `getProtocolDescription()` - описание протоколов
- `getPortPurpose()` - назначение портов

---

## Алгоритмы и методы

### Обнаружение активных хостов

**Алгоритм:**
1. Генерация списка IP адресов из диапазона сети
2. Параллельная проверка доступности каждого IP
3. Проверка популярных портов: 80, 443, 22, 135, 139, 445
4. Если хотя бы один порт отвечает - хост считается активным

**Реализация:**
```go
func (ns *NetworkScanner) isHostAlive(ip string) bool {
    commonPorts := []string{"80", "443", "22", "135", "139", "445"}
    for _, port := range commonPorts {
        conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, port), ns.timeout)
        if err == nil {
            conn.Close()
            return true
        }
    }
    return false
}
```

**Особенности:**
- Использует TCP connect вместо ICMP ping (работает через firewall)
- Проверяет несколько портов для надежности
- Учитывает контекст для возможности отмены

### Сканирование портов

**Алгоритм:**
1. Для каждого активного хоста
2. Параллельно проверяются все указанные порты
3. Используется TCP connect для проверки
4. Результаты сохраняются в структуре ScanResult

**Реализация:**
```go
func isPortOpen(host string, port int, timeout time.Duration) bool {
    address := fmt.Sprintf("%s:%d", host, port)
    conn, err := net.DialTimeout("tcp", address, timeout)
    if err != nil {
        return false
    }
    defer conn.Close()
    return true
}
```

**Особенности:**
- TCP connect сканирование для пользовательского диапазона портов
- Поддержка настраиваемого таймаута
- Опциональная проверка популярных UDP-портов (`--udp`)

### Построение топологии: стратегия связей

**Источник и уверенность связи:**
- LLDP -> `SourceType=lldp`, `Confidence=high`
- FDB/MAC -> `SourceType=fdb`, `Confidence=medium`
- Выводные связи -> `SourceType=inferred`, `Confidence=low` (если используются)

**Правила дедупликации:**
1. Сначала добавляются LLDP-связи.
2. FDB-связи добавляются только если не проигрывают по confidence уже найденной связи.
3. Для одной пары endpoint сохраняется наиболее достоверная связь.
4. При `partial SNMP` для endpoint confidence автоматически понижается (`high→medium`, `medium→low`).

**Фильтрация MAC в FDB:**
- Игнорируются:
  - `ff:ff:ff:ff:ff:ff` (broadcast)
  - multicast (I/G bit первого байта)
  - `00:00:00:00:00:00`
  - MAC самого коммутатора

### Обработка частичных SNMP отказов

**Принцип:** partial failure не должен ломать построение топологии.

- Критичная ошибка: не удалось подключиться ни с одной community (`connect_error`).
- Некритичная ошибка: отказ по одному из OID (`query_error`) при успешном подключении.
- В CLI/GUI выводится summary:
  - сколько целей,
  - сколько успешных/partial/failed,
  - детали отказов.

### Получение MAC адресов

**Алгоритм:**
1. Попытка прочитать из системной ARP таблицы (не реализовано)
2. Отправка ARP запроса через pcap
3. Ожидание ARP ответа
4. Извлечение MAC адреса из ответа

**Реализация:**
- Использует библиотеку `gopacket` для работы с ARP
- Требует права администратора на некоторых системах
- Работает только для устройств в той же подсети

**Ограничения:**
- Может не работать без прав администратора
- Медленнее чем TCP connect
- Не работает для устройств в других подсетях

### Определение типа устройства

**Алгоритм:**
Анализ открытых портов и протоколов:

```go
// Роутер/сетевое оборудование: порты 80/443/8080 + 22
// Веб-сервер: порты 80/443/8080/8443
// База данных: порты 3306/5432/1433
// Windows: порты 3389/445
// Linux/Unix: порт 22
// Принтер: порты 9100/515
// IoT: мало портов (< 3)
```

**Особенности:**
- Эвристический подход
- Может быть неточным для нестандартных конфигураций
- Можно улучшить добавлением базы данных устройств

---

## Зависимости

### Основные зависимости

```go
require (
    github.com/google/gopacket v1.1.19       // Сетевые пакеты (ARP)
    github.com/gosnmp/gosnmp v1.x            // SNMP
    github.com/jedib0t/go-pretty/v6 v6.5.4   // Таблицы
    fyne.io/fyne/v2 v2.x                     // GUI
)
```

### Косвенные зависимости

```go
require (
    github.com/mattn/go-runewidth v0.0.15    // Ширина символов
    github.com/rivo/uniseg v0.4.4            // Unicode сегментация
    golang.org/x/net v0.19.0                 // Сетевые утилиты
    golang.org/x/sys v0.15.0                 // Системные вызовы
)
```

### Системные требования

- **Go:** 1.24 или выше
- **Права:** Для MAC адресов может потребоваться root/sudo
- **Сеть:** Локальная сеть для сканирования

---

## Производительность

Актуальный baseline и perf budget для `Этап 1 / P3` зафиксирован в:
- `docs/P3_PERF_BASELINE.md`

### Оптимизации

1. **Параллельное сканирование:**
   - Использует горутины для параллельной работы
   - Настраиваемое количество потоков (по умолчанию 50)

2. **Двухэтапное сканирование:**
   - Сначала проверка доступности хостов
   - Затем сканирование портов только на активных хостах

3. **Ограничение общей параллельности порт-сканирования:**
   - На уровне `scanner` используется адаптивный лимит проверок портов на хост,
     рассчитываемый из общего budget и числа host-worker.
   - Это предотвращает всплески нагрузки вида `hosts × 100` и снижает риск деградации UI/сети на больших диапазонах.

4. **GUI автопрофиль для крупных подсетей:**
   - В `startScan()` применяется мягкая автокоррекция `threads` и диапазона портов для больших CIDR.
   - Поведение управляется чекбоксом `Автопрофиль сканирования (рекомендуется)` и может быть отключено пользователем.
   - Для рекомендованных настроек GUI сохраняет отдельный класс профиля (`small/medium/large/very-large`) в `Preferences` и восстанавливает бейдж профиля при следующем запуске.
   - Для обратной совместимости сохранен fallback на legacy-текст бейджа, если класс профиля отсутствует.

3. **Эффективное использование памяти:**
   - Результаты сохраняются только для активных хостов
   - Использование указателей и слайсов

### Производительность

**Типичные показатели:**
- Проверка хоста: ~100-300ms (зависит от таймаута)
- Сканирование порта: ~50-200ms
- Сканирование сети /24 (254 хоста, 1000 портов): ~5-15 минут

**Факторы влияния:**
- Размер сети
- Количество портов
- Таймаут подключения
- Количество потоков
- Скорость сети

### Рекомендации по производительности

1. **Для быстрого сканирования:**
   ```bash
   --ports 80,443,22 --timeout 1 --threads 200
   ```

2. **Для точного сканирования:**
   ```bash
   --ports 1-1000 --timeout 5 --threads 100
   ```

3. **Для полного сканирования:**
   ```bash
   --ports 1-65535 --timeout 3 --threads 50
   ```

---

## Ограничения

### Технические ограничения

1. **UDP-сканирование ограничено:**
   - При `--udp` проверяется набор популярных UDP-портов
   - Полноценный массовый UDP-скан в произвольном диапазоне не реализован

2. **MAC адреса:**
   - Могут не определяться без прав администратора
   - Работают только в той же подсети
   - Требуют библиотеку pcap

3. **Обнаружение хостов:**
   - Использует TCP connect, не ICMP ping
   - Может пропустить устройства без открытых портов
   - Зависит от firewall настроек

4. **Инструменты Ping/Traceroute:**
   - Используют внешние системные команды (`ping`, `tracert`/`traceroute`)
   - Требуют наличие этих утилит в `PATH`
   - Формат парсинга зависит от локали и версии системной утилиты

5. **DNS lookup:**
   - По умолчанию использует resolver Go/ОС
   - Для кастомного DNS-сервера применяется explicit resolver (`--dns-server`)
   - При сетевых ограничениях/фильтрации возможны timeout и частичные ответы

6. **WOL (Wake-on-LAN):**
   - Обычно работает только в пределах L2-сегмента
   - Для межсегментного сценария требуется directed broadcast/relay
   - Устройства должны поддерживать и иметь включенный WOL

7. **Баннеры/версии сервисов:**
   - Сбор баннеров может замедлять сканирование
   - Для части служб корректно определяется только `version` или `нет ответа`
   - HTTPS-баннеры читаются в режиме best-effort (без строгой проверки сертификата)

8. **Определение устройств и ОС:**
   - Эвристический подход, может быть неточным
   - Ограниченная база производителей по MAC
   - Active-режим определения ОС использует дополнительные портовые сигнатуры и может давать ложноположительные срабатывания

9. **Remote Exec (P3 MVP):**
   - Требуется явное подтверждение `--remote-exec-consent I_UNDERSTAND`
   - Требуется allowlist хостов и команд (`--remote-exec-policy-file` или `--remote-exec-allow-*`)
   - Рекомендуется `--remote-exec-policy-strict` для production-сценариев
   - В policy запрещен wildcard `*` для хостов/команд
   - В CLI/audit применяется маскирование типовых secret-паттернов (`password/token/secret/api-key`)
   - По умолчанию включен режим `dry-run` (`--remote-exec-dry-run=true`)
   - Для `wmi` и `winrm` поддержка только на Windows
   - Каждая операция журналируется в JSONL (`--remote-exec-audit-log`)
   - Для security report доступен переключатель `--security-report-redact` (default `true`), позволяющий отключить маскирование только для отладки
   - Для `--security-report-redact=false` требуется explicit consent: `--security-report-unsafe-consent I_UNDERSTAND_UNREDACTED_REPORT`
   - В HTML security report явно фиксируется статус: `REDACTION: ON|OFF`
   - Поддержан `--security-report-file auto` с автоименованием: `security-report-redacted-<report-id>.html` / `security-report-unredacted-<report-id>.html`
   - В HTML security report фиксируется metadata: режим генерации (`auto/manual`), версия policy (`v1`) и факт использования unsafe-consent (`yes/no`)

10. **Risk Signatures (Stage2 P2):**
   - Сигнатуры локальные и эвристические; не заменяют полноценный CVE scanner.
   - Качество findings зависит от полноты скана (`ports/service/banner/device-type`).
   - Для части сигнатур рекомендуется включать `--grab-banners`.

11. **Device Control (Stage2 P2):**
   - Поддерживаются только явные действия `status`/`reboot` по заданному URL.
   - Для `reboot` требуется явное подтверждение `--device-confirm I_UNDERSTAND`.
   - Реализованы ограниченные API-профили: `generic-http`, `tp-link-http`.
   - Все действия журналируются в JSONL (`--audit-log`).
   - Функция предназначена только для собственных/разрешенных устройств и сетей.

### Ограничения и права по ОС (P3)

| Функция | Windows | macOS | Linux | Примечание |
|---|---|---|---|---|
| `ping` (инструменты) | Требуется `ping` в `PATH` | Требуется `ping` в `PATH` | Требуется `ping` в `PATH` | При отсутствии утилиты возвращается `not_installed` |
| `traceroute/tracert` | Используется `tracert` | Используется `traceroute` | Используется `traceroute` | Формат raw-вывода зависит от локали/версии утилиты |
| DNS lookup | Через Go resolver | Через Go resolver | Через Go resolver | `--dns-server` использует explicit resolver |
| Whois | Требуется `whois` в `PATH` (часто отсутствует по умолчанию) | Требуется `whois` в `PATH` | Требуется `whois` в `PATH` | При отсутствии возвращается `not_installed` |
| Wi-Fi info | `netsh wlan show interfaces` | `airport -I` | `nmcli ... dev wifi list` | Инструмент OS-specific; при отсутствии утилиты возвращается `not_installed` |
| MAC через ARP | Может требовать повышенные права | Может требовать повышенные права | Может требовать повышенные права | Без прав часть MAC может быть недоступна |
| WOL (`--wol-mac`) | Работает при доступном UDP broadcast | Работает при доступном UDP broadcast | Работает при доступном UDP broadcast | Обычно в пределах L2-сегмента |
| Device Control (`--device-action`) | HTTP endpoint + опц. Basic Auth | HTTP endpoint + опц. Basic Auth | HTTP endpoint + опц. Basic Auth | Для `reboot` требуется `--device-confirm I_UNDERSTAND` |

### Диагностика ошибок execution-слоя (P3)

Инструменты (`ping`, `traceroute`, `dns`, `whois`, `wifi`) нормализуют ошибки в единые коды:

- `not_installed` — внешняя утилита не найдена в `PATH`.
- `permission_denied` — недостаточно прав для запуска или сетевой операции.
- `timeout` — превышен таймаут выполнения (помогает увеличение `--tool-timeout`).
- `network_error` — ошибка сети/маршрутизации/DNS на этапе выполнения.
- `parse_error` — ошибка разбора/нормализации вывода (зарезервировано).
- `unknown` — прочие ошибки выполнения.

Примечание по `wifi`: после выполнения OS-утилиты вывод нормализуется в краткую сводку (ключевые поля подключения), и дополнительно возвращается `Raw output` для диагностики и кросс-проверки.

Дополнительно по устойчивости парсинга:
- Linux (`nmcli -t`): поддерживается разбор escaped-разделителей в SSID (например, `Office\:Guest` -> `Office:Guest`).
- Windows (`netsh wlan show interfaces`): поддерживаются локализованные RU-ключи (`Имя`, `Состояние`, `Сигнал`, `Канал`, `Скорость приема/передачи`, `Проверка подлинности`, `Тип радиомодуля`).
- Состояние Wi-Fi нормализуется в `connected` / `disconnected`, при отсутствии поля устанавливается `unknown`.

### Функциональные ограничения

1. **Точность топологии зависит от SNMP:**
   - При недоступном SNMP топология может быть неполной
   - LLDP/FDB зависят от вендора и конфигурации устройства

2. **Фильтрация реализована, но остается пространство для расширения:**
   - В GUI уже есть фильтры: query/type/open-only/CIDR/port-state, presets, sort.
   - На текущем этапе отсутствуют более сложные составные фильтры (например, расширенные правила по сервисам/протоколам/сегментам).

3. **Нет планировщика:**
   - Нет автоматического периодического сканирования
   - Нет сравнения результатов между временными срезами

---

## Будущие улучшения

### Краткосрочные (легко реализуемые)

1. **Улучшение качества топологии:**
   - Дополнительные источники соседства (CDP, vendor-specific MIBs)
   - Улучшенная корреляция multi-link/trunk сценариев
   - Визуальная подсветка confidence в GUI/экспорте

2. **Улучшенная фильтрация:**
   - Фильтр по типу устройства
   - Фильтр по протоколам
   - Фильтр по портам

3. **Чтение ARP таблицы:**
   - Реализация для Linux (/proc/net/arp)
   - Реализация для macOS (arp -a)
   - Реализация для Windows

### Среднесрочные (требуют больше работы)

1. **Расширение UDP:**
   - Поддержка произвольных диапазонов UDP портов
   - Более точная эвристика открытости UDP

2. **Улучшенное определение устройств:**
   - База данных устройств
   - Fingerprinting по ответам
   - Определение версий сервисов

3. **Веб-интерфейс:**
   - REST API
   - Веб-интерфейс для просмотра результатов
   - История сканирований

### Долгосрочные (большие изменения)

1. **Распределенное сканирование:**
   - Сканирование с нескольких точек
   - Координация результатов

2. **Анализ безопасности:**
   - Обнаружение уязвимостей
   - Рекомендации по безопасности
   - Отчеты о безопасности

3. **Интеграции:**
   - Интеграция с системами мониторинга
   - Уведомления об изменениях
   - API для внешних систем

---

## Безопасность

### Соображения безопасности

1. **Сетевые запросы:**
   - Может быть обнаружено системами IDS/IPS
   - Может вызвать подозрения в корпоративных сетях

2. **Права доступа:**
   - Требует прав администратора для MAC адресов
   - Может требовать специальных разрешений

3. **Этичное использование:**
   - Используйте только в своих сетях
   - Получайте разрешение перед сканированием
   - Соблюдайте политики безопасности

### Рекомендации

- Используйте только в тестовых/собственных сетях
- Получайте письменное разрешение перед сканированием корпоративных сетей
- Информируйте администраторов сети о сканировании
- Соблюдайте законы и политики безопасности

---

## Дополнительные ресурсы

- [Инструкция по эксплуатации](../Инструкция%20по%20эксплуатации.md) - Полная инструкция по эксплуатации (русский язык)
- [Руководство пользователя](USER_GUIDE.md) - Подробное руководство пользователя
- [Архитектура проекта](ARCHITECTURE.md) - Описание архитектуры проекта
- [GUI документация](GUI.md) - Документация по GUI версии
- [Инструкция по установке](INSTALL.md) - Инструкции по установке
- [README.md](../README.md) - Основная документация проекта

---

## Регрессионные smoke-проверки

Для быстрых проверок CLI путей используйте:

```bash
./scripts/smoke-cli-no-topology.sh
./scripts/smoke-cli-topology.sh
./scripts/smoke-cli-tools.sh
```

```powershell
.\scripts\smoke-cli-no-topology.ps1
.\scripts\smoke-cli-topology.ps1
.\scripts\smoke-cli-tools.ps1
```

- `no-topology` проверяет, что без `--topology` не появляется SNMP summary.
- `topology` проверяет наличие SNMP summary в режиме топологии.
- `tools` проверяет ключевые tool-режимы (`--ping`, `--dns`, `--whois`, `--wifi`) и секции raw-вывода.
- `smoke-d-track-topology-export` проверяет экспорт `json/graphml/png` и эквивалентность множеств узлов/связей между `json` и `graphml` (с fallback для `png`, если `dot` недоступен).

```bash
./scripts/smoke-d-track-topology-export.sh
```

```powershell
.\scripts\smoke-d-track-topology-export.ps1
```

Для ручной проверки совместимости `GraphML` во внешних инструментах используйте:

- `docs/GRAPHML_COMPATIBILITY_CHECK.md`

### Closure-проверки этапов

Для формального закрытия этапов используйте агрегирующие скрипты:

```bash
# Linux/macOS (6-command runbook, copy/paste)
./scripts/p1-closure-check.sh && ./scripts/p2-closure-check.sh && ./scripts/p3-closure-check.sh && ./scripts/stage2-p1-closure-check.sh && ./scripts/stage2-p2-closure-check.sh && ./scripts/stage2-p3-closure-check.sh
```

```powershell
# Windows PowerShell (6-command runbook, copy/paste)
.\scripts\p1-closure-check.ps1; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }; .\scripts\p2-closure-check.ps1; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }; .\scripts\p3-closure-check.ps1; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }; .\scripts\stage2-p1-closure-check.ps1; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }; .\scripts\stage2-p2-closure-check.ps1; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }; .\scripts\stage2-p3-closure-check.ps1
```

```bash
# Этап P1
./scripts/p1-closure-check.sh

# Этап P2
./scripts/p2-closure-check.sh

# Этап P3
./scripts/p3-closure-check.sh

# Этап Stage2/P1
./scripts/stage2-p1-closure-check.sh

# Этап Stage2/P2
./scripts/stage2-p2-closure-check.sh

# Этап Stage2/P3
./scripts/stage2-p3-closure-check.sh
```

```powershell
# Этап P1
.\scripts\p1-closure-check.ps1

# Этап P2
.\scripts\p2-closure-check.ps1

# Этап P3
.\scripts\p3-closure-check.ps1

# Этап Stage2/P1
.\scripts\stage2-p1-closure-check.ps1

# Этап Stage2/P2
.\scripts\stage2-p2-closure-check.ps1

# Этап Stage2/P3
.\scripts\stage2-p3-closure-check.ps1
```

`p2-closure-check` дополнительно валидирует P2-флаги CLI (`--grab-banners`, `--show-raw-banners`, `--os-detect-active`, `--risk-signatures`, `--device-action`) и ожидаемое неуспешное завершение WOL/device-control в negative-case сценариях.
`stage2-p1-closure-check` дополнительно включает `go test ./cmd/network-scanner -run Whois`, что фиксирует e2e-путь `--whois` в CLI (`runToolsMode`) с RDAP fallback при отсутствии системного `whois`.
`stage2-p2-closure-check` дополнительно проверяет генерацию `security report` и наличие секций `CVE Findings` + `Risk Signature Findings`.
`stage2-p3-closure-check` дополнительно проверяет guardrails Stage2/P3: redaction/consent/report-id/auto filename для `security report` и policy/allowlist ограничения для `remote-exec`.
CI helper-скрипты `check-ci-status.*` и `trigger-ci-workflow.*` работают в strict-режиме: required jobs включают `Lint`, `Test*`, `Build and Smoke*`, `Stage2 P1 Closure`, `Stage2 P3 Closure`.

---

**Версия документа:** 1.0.5  
**Последнее обновление:** 2026-04-23

