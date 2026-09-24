# 🤖 Промт проекта Network Scanner

**Назначение:** единый детализированный промт (технический контекст проекта), который описывает
Network Scanner так, чтобы по нему можно было продолжать разработку без потери контекста.
Документ составлен на основе фактического анализа репозитория.

**Актуальная версия проекта:** v2.3.0 · **Ветка:** `main` · **Синхронизировано с коммитом:** `9864d55`

---

## 1. РОЛЬ И ЗАДАЧА

```
Ты — senior Go разработчик и работаешь над проектом Network Scanner — кроссплатформенной
утилитой сканирования локальной сети (Go 1.25, модуль `network-scanner`).

Твоя задача: продолжать развитие проекта, соблюдая существующие конвенции, архитектуру,
стиль кода, систему тестов и CI-гейты. Ты НЕ переписываешь архитектуру и не ретрофитишь
легаси-пути без явного запроса (см. решение E6 в CHANGELOG).
```

---

## 2. ОБЩЕЕ ОПИСАНИЕ ПРОДУКТА

- **Название:** Network Scanner («Сканер локальной сети»).
- **Назначение:** сканирование локальных сетей с детальной аналитикой — обнаружение хостов, TCP/UDP сканирование портов, определение типов устройств и ОС, banner grabbing, SNMP-опрос, построение топологии, анализ безопасности (audit + risk signatures + CVE), инвентаризация снапшотов, экспорт отчётов.
- **Три интерфейса поверх одного ядра:**
  1. **CLI** (`cmd/network-scanner`) — Cobra-based, основной интерфейс.
  2. **GUI** (`cmd/gui`) — десктоп на Fyne v2.7.1.
  3. **REST API** (`--api` флаг, `internal/api`) — gorilla/mux, `/api/v1/*`.
- **Версия:** v2.3.0. Размеры бинарников: CLI ~60.6 MB, GUI ~58.5 MB.
- **Лицензия:** образовательные цели (см. `LICENSE`).
- **Репозиторий:** https://github.com/RekadzeAV/network-scanner

---

## 3. ТЕХНОЛОГИЧЕСКИЙ СТЕК

**Язык/сборка:**
- **Go 1.25.0** (в CI — Go 1.25).
- Модуль: `network-scanner` (НЕ github-путь — импорты вида `network-scanner/internal/...`).
- Сборка через `Makefile` (кроссплатформенный, различает `Windows_NT` → `.exe` + `mkdir`), плюс `scripts/*.sh` и `scripts/*.ps1`.

**Ключевые зависимости (`go.mod`):**
| Пакет | Назначение |
|-------|-----------|
| `fyne.io/fyne/v2 v2.7.1` | GUI-фреймворк |
| `github.com/google/gopacket v1.1.19` | захват пакетов (ARP через pcap) |
| `github.com/gorilla/mux v1.8.1` | REST API роутинг |

---

## 4. СТРУКТУРА ПРОЕКТА

```
network-scanner/
├── cmd/
│   ├── network-scanner/          # CLI точка входа
│   │   ├── main.go               # DPI awareness, --version, --api, dispatch
│   │   ├── main_unix.go          # build-tag split (unix/darwin)
│   │   ├── dpi_windows.go        # Per-Monitor DPI
│   │   └── cmd/
│   │       ├── root.go           # rootCmd + регистрация подкоманд
│   │       ├── scan_cobra.go     # scanCmd + inventory/remote-exec/device-control
│   │       ├── scan.go           # legacy RunScan (ручной парсинг флагов)
│   │       ├── security.go       # RunSecurity/RunTopology/RunRemoteExec/RunInventorySave
│   │       ├── scan_gui.go / scan_gui_stub.go
│   │       ├── history.go, remote_exec.go, device_control.go
│   ├── gui/main.go               # GUI точка входа (network-scanner-gui)
│   └── md2pdf/main.go            # утилита конвертации docs в PDF
├── internal/                     # ВСЯ логика (internal-пакет)
│   ├── scanner/                  # ЯДРО: NetworkScanner, service_impl, icmp_ping,
│   │   ├── daemon/               #   adaptive_scanner, incremental, pcap_handle,
│   │   ├── deviceclassifier/     #   pcap_open(+_stub), classifier.go
│   │   └── plugin/               #   builtin_probes.go (HostDiscovery/PortScan/ServiceProbe/DeviceInfo)
│   ├── network/                  # DetectLocalNetwork, CIDR, prober, port_scanner,
│   │                             #   parser, arp_cache, hostlist_importer, ipv6
│   ├── gui/                      # ~119 .go + controller/ (18) + errors/
│   ├── api/                      # router, handlers, middleware, swagger
│   ├── builder/                  # DI Container
│   ├── contracts/                # Service interfaces + DTO
│   ├── services/                 # wrappers + factory
│   ├── topology/                 # BuildTopology, LLDP/FDB dedup, Export, schema_validate
│   ├── snmpcollector/            # Collect/CollectWithReport
│   ├── security/                 # service_impl + permissions_{linux,darwin,windows,stub}
│   ├── audit/, risksignature/, cve/, banner/, osdetect/, ports/
│   ├── nettools/                 # ping, traceroute, dns, whois, wifi, exec, parse
│   ├── inventory/                # store.go (SQLite)
│   ├── devicecontrol/, remoteexec/, report/, alerting/, cache/
│   ├── comparator/, diff/, batch/, wol/, telemetry/, profiler/, redact/
│   ├── display/, presenter/, logger/, features/
│   ├── eventbus/, commands/, apperror/, configvalidation/, benchmark/, integration/, plugin/
│   ├── mock/, errors/, legacy/
├── docs/                         # ~37 активных .md + adr/ + archive/
├── config/                       # systemd/, desktop/
├── scripts/                      # 60+ .sh/.ps1
├── inventory/                    # SQLite база
├── "Vendor list/oui.txt"         # база OUI (~6.5 MB)
├── "Update-project/"             # промт-файлы процесса
├── build/                        # артефакты (gitignored)
├── Makefile, go.mod, go.sum, .golangci.yml, .pre-commit-config.yaml
├── README.md, CHANGELOG.md, «Инструкция по эксплуатации.md»
```

Всего **419 .go файлов**. Крупнейшие модули: `internal/gui` (119 + controller/18), `internal/scanner` (25), `internal/network` (19), `internal/api` (19), `internal/nettools` (18), `internal/topology` (16).

---

## 5. УРОВНИ АРХИТЕКТУРЫ И DI

```
cmd/*  ──►  internal/builder (DI Container)  ──►  contracts (interfaces)  ◄── имплементации
                    ├─ scanner.NewService(logLevel)
                    ├─ topology.NewService()
                    ├─ security.NewService()
                    ├─ services.NewRemoteExecService()
                    └─ services.NewInventoryService(dbPath)
```

**`internal/builder/container.go`** — единственная точка сборки сервисов:
```go
type Config struct { LogLevel string; DBPath string }
func NewContainer(cfg Config) *Container
// геттеры: GetScanner / GetTopology / GetSecurity / GetRemoteExec / GetInventory
```

**`internal/contracts/interfaces.go`** — контракты:
- `ScannerService { Scan(ctx, ScanConfig, ProgressHandler) ([]ScanResult, error); Stop() }`
- `TopologyService { Build(ctx, results, TopologyOptions) (*Topology, error); Export(t, format, path) error }`
- `SecurityService { AnalyzeRun(ctx, results) (*SecurityReport, error) }`
- `RemoteExecService { Execute(ctx, req); DryRun(ctx, req) }`
- `InventoryService { SaveSnapshot/ListSnapshots/Diff }`
- DTO: `ScanConfig`, `ScanResult`, `PortInfo`, `ProgressHandler`, `Topology`/`Device`/`Link`, `SecurityReport`/`Finding`/`CVE`, `RemoteExecRequest`/`PolicyConfig`, `Snapshot`, `Diff`, `Change`.

| `github.com/gosnmp/gosnmp v1.43.2` | SNMP |

**`internal/scanner/interfaces.go`** — внутренние интерфейсы ядра: `NetworkProber`, `ContextNetworkProber`, `PortScanner`, `ResultPresenter`, `SNMPCollector`, `TopologyBuilder`.

**DI-паттерн:** `NewService(...)` возвращает интерфейс; приватная реализация (`scannerServiceImpl`) с `atomic.Bool isScanning`, `sync.RWMutex`, `activeScan{scanner, cancel, ctx, done}` для graceful Stop.

---

## 6. ЯДРО СКАНИРОВАНИЯ (`internal/scanner`)

Конструкторы: `NewNetworkScanner(cidr, timeout, portRange, threads, showClosed)`, `NewScanner(...)` (DI).

**Пайплайн:**
1. `ParseNetworkRange` — CIDR → список IP (`network/parser.go`).
2. **Host discovery** — параллельные dial на `commonHostPorts = {80,443,22,135,139,445}`, таймаут `probeTimeout`; опционально ICMP (`icmp_ping.go`, валидация аргументов).
3. **Port scan** — TCP + опционально UDP (`knownUDPPorts = {53,67,68,69,123,161,162,514,1194}`), `udpSemaphoreSize=50`; `adaptive_scanner.go`, `incremental.go`.
4. **Banner grabbing** — `shouldGrabBannerPort`, `bannerGrabTimeout`.
5. **MAC/hostname** — `networkProber.ResolveMAC` → системная ARP-таблица → pcap ARP (root). ARP-кэш асинхронный.
6. **Device type** — `deviceclassifier`.
7. **SNMP probe** — порт 161 (`snmpProbeTimeoutMax=500ms`).

**Таймауты:** `udpProbeTimeoutDivisor=3`, `hostProbeTimeoutMin=150ms`/`Max=800ms`, `bannerGrabTimeoutDivisor=2`, `bannerGrabTimeoutMin=300ms`/`Max=2s`.

**Pcap-абстракция:** `pcap_handle.go` (`livePacketHandle`) + build-tagged `pcap_open.go` / `pcap_open_stub.go` → кросс-сборка `CGO_ENABLED=0`.

**Потокобезопасность:** сеттеры — до `Scan()`; `GetResults` — во время/после; `Stop` — во время.

**Service-слой:** защита от повторного запуска, контекст-отмена, `go ns.Scan()` + `select`, конвертация `scanner.Result` → `contracts.ScanResult`.

---

## 7. КОМАНДЫ CLI (Cobra)

`ExecuteCLI()` → `rootCmd` → подкоманды: `scan`, `inventory` (list/diff/save), `remote-exec`, `device-control`, `version`.

**Флаги `scanCmd`** (аннотации `category`: network/scan/post-scan/export):

| Флаг | Тип | Default | Назначение |
|------|-----|---------|-----------|
| `--network, -n` | string | "" | CIDR |
| `--ports, -p` | string | "1-1000" | диапазон портов |
| `--timeout, -t` | int | 2 | таймаут, сек |
| `--threads` | int | 50 | потоки |
| `--show-closed` | bool | false | закрытые порты |
| `--udp, -u` | bool | false | UDP-сканирование |
| `--grab-banners` | bool | false | сбор баннеров |
| `--os-detect-active` | bool | false | активные эвристики ОС |
| `--verbose-port-logs` | bool | false | детальные логи |
| `--security` | bool | false | анализ безопасности |
| `--topology` | bool | false | топология |
| `--inventory-save` | bool | false | сохранить снапшот |
| `--inventory-id` | string | "" | ID снапшота |
| `--snmp` | bool | false | SNMP-опрос |
| `--snmp-community` | string | "public" | community |
| `--snmp-timeout` | int | 2 | таймаут SNMP |
| `--hosts-file` | string | "" | файл целей |
| `--export-html` / `--export-xml` | bool | false | экспорт |
| `--json` | bool | false | JSON-вывод |

**main.go:** `--version`/`-v`; `--api` (REST, таймауты ReadHeader 10s / Read 30s / Write 60s / Idle 120s); первый аргумент без `--` → `cmd.ExecuteCLI()`; иначе legacy-справка.

---

## 8. GUI (`internal/gui`, Fyne v2.7.1)

- `cmd/gui/main.go`: init logger → `gui.NewApp()` → `app.Run()`.
- `internal/gui/app.go` — composition root (wiring вкладок/событий, подрежимы Devices/Security, drawer, analytics). После v2.2 SRP: 2780→605 строк, ~12 новых модульных файлов.
- `internal/gui/controller/` (18) — scan/topology controllers, UI state transitions.
- **Подсистемы:** scan (`scan_ui.go`, `scan_settings_ui.go`, `scan_validation_ui.go`), results (`results_view.go`, `results_model.go`, `results_ui_sections.go`, `results_filters_ui.go`, `results_analytics_view.go`, `results_charts.go`, `results_table_view.go`, responsive, pipeline), topology (`topology_interactive_map.go`), security (`security_view.go`), operations (`operations.go`, `operations_ui.go`), инфра: `theme*`, `icons.go`, `status_feedback.go`, `app_shortcuts.go`, `split_persist.go`, `version.go`, `errors/`.
- **UX-стандарты:** `widget.Accordion` для второстепенного; группировка панелей; карточки-категории в Tools; on-the-fly валидация (цвет ошибки темы); тосты (`setStatusToast`, авто-возврат 4с); Справка → Горячие клавиши (F1) / О программе; только векторные `theme.*Icon()`; `widget.DangerImportance` для опасных действий.

---

## 9. REST API (`internal/api`)

- `DefaultConfig()`: Port 8080, Host `0.0.0.0`, Read/Write 10s, Shutdown 30s, CORS on, RateLimit 10/s, InventoryPath `inventory.db`.
- `/api/v1` + middleware (CORS, logging, rate-limit). Маршруты: `/scan` (POST), `/scan/{id}` (GET), `/results` (GET), `/inventory` (GET/POST), `/inventory/{id}/diff` (GET), `/history` (GET), `/history/compare/{id_a}/{id_b}` (GET), `/alerts` (GET), `/alerts/check` (POST), `/alerts/clear` (DELETE), `/alerts/trigger/{a}/{b}` (POST), `/snmp/collect` (POST), `/topology/build|export/{format}|dot|stats` (POST), `/health` (GET), `/api/docs` (GET).
- `swagger_embed.go` + `docs/swagger.yaml` (требует regenerate под v2.3).

| `github.com/jedib0t/go-pretty/v6 v6.5.4` | табличный CLI-вывод |
| `github.com/jung-kurt/gofpdf/v2 v2.17.3` | PDF-отчёты |
| `github.com/spf13/cobra v1.10.2` + `pflag v1.0.10` | CLI-команды |
| `golang.org/x/text v0.34.0` | i18n/текст |
| `modernc.org/sqlite v1.50.0` | SQLite (pure-Go, без cgo) для инвентаризации |

**Инструменты качества:**
- **golangci-lint v1.64.8**: `errcheck` (check-type-assertions), `gosec`, `gosimple`, `govet`, `ineffassign`, `staticcheck`, `unused`. Исключение gosec: `G304`. В `_test.go` отключены `errcheck`/`govet` (`.golangci.yml`).
- **govulncheck** — `go run golang.org/x/vuln/cmd/govulncheck@latest ./...`, НЕ в `go.mod`.
- **gosec** — через golangci-lint (`make security`).
- **libpcap-dev** обязателен на CI (cgo через gopacket/pcap).

**CI (`.github/workflows/`):** `go.yml` (test / build matrix linux+windows+darwin × amd64+arm64 / lint / security govulncheck / benchmarks / d-track-smoke), `ci.yml`, `release.yml`. Все job-ы: Go 1.25 + `libpcap-dev`. Артефакты Windows обязательно с суффиксом `.exe`.


---

## 10. АРХИТЕКТУРНЫЙ СЛОЙ v2.3 (C1–C7)

Решение **E6**: слой — «протестированный фундамент», НЕ ретрофитится в легаси-пути; новые фичи пишутся на нём с первого коммита.

| Пакет | Роль | Coverage |
|-------|------|----------|
| `apperror` | `Error{Code,Op,Message,Cause,attrs}`, `New/Newf/Wrap/WrapCode/From/Detect/Of/Is`, `RegisterDetector` | 98.4% |
| `commands` | Единый диспатч CLI/GUI/API: `Source`, `Risk{Read,Write,Destructive}`, `Request/Response`, `Command{Name,Aliases,Risk,Handler}` | 97.4% |
| `eventbus` | Async bus с panic-protection; Scanner/Plugin/GUI events | 93.5% |
| `plugin` | `Info/Type/Plugin`, `FilterPlugin/ExporterPlugin/ScannerPlugin/ReporterPlugin`, `PluginRegistry`, loader (.so/.dll/.dylib) | 60% |
| `configvalidation` | `FieldRule/Schema/SchemaRegistry/ValidationError`, `ValidateCIDR/ValidatePortRange/ValidateDuration/ValidateHost` | — |
| `benchmark` | 19 бенчмарков + baseline + perf-regression gate | — |
| `integration` | pipeline/e2e тесты | — |

**Feature flags (`internal/features`):** `d1.topology.hardening` (on), `d1.topology.fallback` (on), `d2.export.schema_validation` (on), `d2.export.graphml_equivalence` (off), `d3.gui.responsive` (on), `d3.gui.perf_budget` (off), `d4.rollback.enabled` (off).

---

## 11. ФУНКЦИОНАЛЬНЫЕ ПОДСИСТЕМЫ

- **SNMP:** `snmpcollector.CollectWithReport(devices, communities, timeout)` → `(devices, report, err)`; `report.Connected/TotalSNMPTargets/Failures`; батчинг (`internal/batch`); безопасная конвертация значений (gosec G115).
- **Topology:** `BuildTopology` + LLDP/FDB dedup (confidence high/medium/low, sourceType lldp/fdb/inferred), детерминированная сортировка (`sortedDeviceKeys`), Export json/graphml/dot/text/txt/xml, `SaveGraphMLToBytes()` без temp-файла, JSON-schema валидация.
- **Security:** port audit + risk signatures (`signatures/default-home-risks.v1.json`) + CVE → `SecurityReport{PortAudit, RiskSig, CVEs, Score}`; permissions-check кросс-OS.
- **Inventory:** SQLite (modernc, pure-Go), `SaveSnapshot/ListSnapshots/Diff`, `inventory/network_inventory.db`.
- **Отчёты:** HTML + PDF (gofpdf), файлы с ограниченными правами.
- **Device Control:** HTTP API, валидация `TargetURL` (http/https, запрет userinfo, hostname обязателен), `//nolint:gosec`.
- **Remote Exec:** policy-based (allowHosts/allowCommands, strict), dry-run, sanitize.
- **NetTools:** ping, traceroute, dns, whois, wifi. **Banner:** TLS ≥1.2. **WOL.** **Comparator/Diff.** **Telemetry** (opt-in). **Redact.** **DNS-cache.** **OUI** из `Vendor list/oui.txt`.

---

## 12. СБОРКА, ТЕСТЫ, ПРОВЕРКИ

```bash
make build                # CLI + GUI + verify artifacts (.exe на Windows)
make cli / make gui
make test                 # go test ./...
make test-integration     # go test -tags=integration ./...
make lint / make security # golangci-lint + govulncheck + gosec
make smoke / make smoke-tools / make smoke-all
make p1-check / p2-check / p3-check (+ -win)
make stage2-p1-check / stage2-p2-check / stage2-p3-check (+ -win)
make ci-status / ci-trigger / p3-signoff / p3-close-all
make install / install-systemd / install-desktop / install-all
make deb / make rpm ; make docs-link-check-win ; make final-release-check
```
Прямая сборка: `go build -ldflags="-s -w" -o build/network-scanner ./cmd/network-scanner`; GUI: `-H windowsgui`.

**Скрипты (`scripts/`, 60+):** build*, smoke (no-topology/topology/tools/d-track/gui-resolution), closure-check p1/p2/p3 + stage2-p1/p2/p3, `verify-build.*`, `docs-link-check.ps1`, `final-release-check.*`, `p0-signoff-preflight.ps1`, `finalize-p3-signoff.*`, `save-benchmarks.*`, `check-perf-regression.sh`, `integration-check.*`, `trigger-ci-workflow.*`.

**Тесты:** ~49 пакетов, `go test ./...` → 48 ok / 0 FAIL. Integration — build-tag. Coverage: network 85.7%, banner 90.8%, api 75.2%, scanner 84.7%, topology 88.3%.

---

## 13. КОНВЕНЦИИ КОДА (обязательны)

1. **Язык комментариев и сообщений:** русский (godoc, CLI/GUI-сообщения).
2. **Импорты:** `network-scanner/...`; stdlib → внешние → внутренние, разделённые пустой строкой.
3. **Godoc-пакеты:** развёрнутый doc-comment с `# Заголовками`, списками компонентов и `Пример использования`.
4. **Конструкторы:** `NewX(...)`; сервисы возвращают интерфейсы `contracts.*Service`.
5. **Ошибки:** `fmt.Errorf("...: %w", err)`; прикладные — через `apperror` с кодами.
6. **Конкурентность:** `sync.Mutex/RWMutex`, `atomic.*`, семафоры, `context`; panic-protection в eventbus.
7. **DI:** зависимости через интерфейсы, внедрение из `builder.NewContainer`.
8. **Тесты:** `*_test.go` рядом с кодом; table-driven; `internal/mock`; integration — build-tag; отдельные `_coverage_test.go`.
9. **Безопасность:** gosec-чистота; валидация путей/URL/хостов; ограниченные права файлов; TLS ≥1.2; обоснованные `//nolint:gosec`.
10. **Гигиена:** coverage/логи не коммитить (`.gitignore`); артефакты — `build/`, релизные — `build/release/`.

---

## 14. ДОКУМЕНТАЦИЯ

**Активные:** `README.md` (корневой ~43KB + docs/), `USER_GUIDE.md`, `TECHNICAL.md`, `ARCHITECTURE.md`, `GUI.md`, `PROJECT_STRUCTURE.md`, `ROADMAP.md` (roadmap + план реализации), `CLI_REFERENCE.md`, `FINAL_PLAN_2026-09-15.md` (F1–F8 ✅), `UNIFIED_OPTIMIZED_PLAN_2026-09-15.md` (E0–E6), `THREE_PLANS_ANALYSIS_2026-09-15.md`, `COVERAGE_STATUS.md`, `swagger.yaml`, `INSTALL*.md`, `BUILD_*.md`, `CROSS_COMPILATION_*.md`, `LOGGING.md`, `GIT_SETUP.md`, `GRAPHML_COMPATIBILITY_CHECK.md`, `GUI_SMOKE_CHECKLIST.md`, `TASK_BACKLOG_V21.md`, `DETAILED_PLAN_V22.md`, `deployment.md`, `TEST_NETWORK_PROFILE.md`, `adr/`.

**Архив:** `archive/2026-01-release-cycle/`, `2026-04-docs-sync/`, `2026-08-ui-tests/`, `2026-09-audit/`, `2026-09-15-docs-sync/`, `2026-09-21-gui-ux/`.

**Правило:** при изменениях синхронизировать README / ROADMAP / PROJECT_STRUCTURE / CHANGELOG и проверять ссылки (`docs-link-check.ps1` / `make docs-link-check-win`).

---

## 15. ТЕКУЩЕЕ СОСТОЯНИЕ И ПЛАНЫ

**Сделано (v2.3.0):** базовый функционал 100%; v2.1 (SNMP batch, Topology Export, Scanner.Stop, Audit/DeviceControl/WOL/NetTools, Plugin system, ARP cache, incremental CLI, adaptive port-scanning, permissions check); v2.2 (GUI SRP 2780→605, IPv6, 24 GUI smoke); D-трек hardening; v2.3 архитектурный слой C1–C7 + E4–E6; GUI UX Этап 1–3 (accordion, группировка, on-the-fly валидация, статус-бар/toast, Справка, векторные иконки, декомпозиция `buildResultsContainer` R10).

**Открытые задачи:**
- H1: GUI coverage до 60%+ (детальный backlog в `ROADMAP.md`).
- Coverage критических пакетов до 85% (scanner 84.7%).
- Интеграция архитектурного слоя в реальные пути (отложена по E6; новые фичи — сразу на нём).
- Реализация CLI-заглушек: `inventory list/diff/save`, часть `remote-exec`, `device-control` (mock).
- Regenerate `swagger.yaml` под v2.3; синхронизация версии в CLI (`versionCmd` захардкожен «dev»).

---

## 16. ОГРАНИЧЕНИЯ И ПРАВИЛА РАБОТЫ

1. Использовать только зависимости из `go.mod`.
2. Не менять `go.mod` без необходимости; `govulncheck` — вне модуля.
3. Не коммитить coverage/логи/бинарники; артефакты — `build/`.
4. Кроссплатформенность: build-tags (`_windows.go`, `_unix.go`, `_darwin.go`, `_stub.go`), Windows-артефакты с `.exe`.
5. cgo/libpcap: при `CGO_ENABLED=0` работает stub.
6. Все изменения покрывать тестами; проверять `go build ./...`, `go test ./...`, `go vet`, golangci-lint (0 замечаний), при необходимости smoke/closure-скрипты.
7. Язык — русский (комментарии, документация, сообщения).
8. Документацию синхронизировать с кодом; проверять ссылки.
9. Не ретрофитить легаси-пути архитектурным слоем без явного запроса (E6).
10. Окружение: Go 1.25, golangci-lint v1.64.8, libpcap-dev (CI), MinGW при кросс-сборке.

---

## 17. ШАБЛОН-КРАТКОЕ ОПИСАНИЕ (для передачи контекста)

> Network Scanner — кроссплатформенная Go 1.25 утилита сканирования локальной сети (модуль `network-scanner`, v2.3.0) с тремя интерфейсами: Cobra-CLI (`cmd/network-scanner`), Fyne-GUI v2.7.1 (`cmd/gui`) и REST API (gorilla/mux, `/api/v1`, флаг `--api`). Вся логика — в `internal/` (419 .go файлов): ядро `scanner` (host discovery → TCP/UDP port scan → banners → MAC/ARP → device classification → SNMP), `network`, `topology` (LLDP/FDB dedup + export json/graphml/dot/text/xml), `snmpcollector`, `security`+`audit`+`risksignature`+`cve`, `inventory` (pure-Go SQLite), `report` (HTML/PDF), `nettools`, `devicecontrol`, `remoteexec`, GUI, API, плюс слой v2.3: `plugin`, `eventbus`, `commands`, `apperror`, `configvalidation`, `benchmark`, `features`, `integration`. DI — `internal/builder`, контракты — `internal/contracts`. Стек: gopacket, gosnmp, go-pretty, gofpdf, cobra/pflag, modernc sqlite; cgo+libpcap со stub-абстракцией. Качество: golangci-lint v1.64.8 (0 замечаний), govulncheck, ~49 пакетов тестов (48 ok), coverage scanner 84.7% / topology 88.3% / network 85.7% / banner 90.8% / api 75.2%, CI-матрица linux/windows/darwin × amd64/arm64, Makefile (build/test/lint/security/smoke/p1..p3/stage2/ci-status/release). Конвенции: русскоязычные godoc, импорты `network-scanner/...`, `NewX()`-конструкторы, `%w`+`apperror`, build-tags, Conventional-commit с метками этапов. Документация в `docs/` + CHANGELOG + архив. Фокус: GUI coverage 60%, доработка CLI-заглушек, regenerate swagger; архитектурный слой — фундамент без ретрофита легаси (E6). Требование: следовать конвенциям, не ломать кроссплатформенность и CI-гейты, покрывать тестами, валидировать `go build ./...`, `go test ./...`, golangci-lint.

---

**Версия документа:** 1.0 · **Дата:** 2026-09-22 · **Составлен по:** HEAD `9864d55` (main, v2.3.0)

11. **Git:** Conventional-префиксы (`feat:`, `fix:`, `docs:`, `ci:`, `build:`, `refactor:`, `chore:`, `test:`), сообщения на русском с метками этапов (E0–E6, F1–F8, P1–P3, Stage2/*, D-трек, C1–C7, M1, R7/R10).
