# Roadmap

Этот файл — **каноническая точка входа** в roadmap проекта и план реализации.
Он объединяет ранее раздельные `ROADMAP.md` и `IMPLEMENTATION_PLAN.md`
(объединение выполнено 2026-09-23; старый план → [archive/2026-09-cycle/IMPLEMENTATION_PLAN_v1.md](archive/2026-09-cycle/IMPLEMENTATION_PLAN_v1.md)).

## Текущее состояние проекта

**Версия:** v2.3.0 (released 2026-09-21)
**Базовый функционал:** ✅ 100% (13/13 задач)
**v2.1 / v2.2 backlogs:** ✅ выполнены
**Архитектурный слой v2.3 (C1–C7):** ✅ пакеты реализованы с тестами; ⚠️ интеграция в приложения — в плане (этап E6)
**Тесты:** ✅ 52 пакета ok (`go test ./... -cover`, прогон 2026-09-24); ⚠️ `internal/scanner/daemon` — плавающий: падает при полном параллельном прогоне, проходит изолированно (3/3)
**Coverage:** scanner 88.1%; api 95.5%; devicecontrol 90.8%; topology 88.3%; apperror 98.4%; commands 97.4%
**Race detector:** ✅ `go test -race ./... -short` — 0 FAIL, 0 DATA RACE (2026-09-24; гонки M2 устранены)
**Дата обновления:** 2026-09-24

## Завершённые этапы (100%)

### Этап 1: Критические исправления
- [x] 1.1 Добавить `main.Version` и переменные сборки
- [x] 1.2 Mock-сервисы для тестирования
- [x] 1.3 Унифицировать error handling

### Этап 2: Функциональные улучшения
- [x] 2.1 REST API (POST /scan, GET /inventory)
- [x] 2.2 Экспорт отчётов (PDF/HTML)
- [x] 2.3 История сканирований (сравнение снапшотов)
- [x] 2.4 Alerting (система уведомлений)

### Этап 3: Продвинутые функции
- [x] 3.1 SNMP сбор данных
- [x] 3.2 Topology discovery (построение карты сети)
- [x] 3.3 GUI интерфейс (Fyne framework, 108 файлов после SRP)

### Этап v2.1: Сервисы и контроллеры
- [x] Пакетная обработка SNMP (BatchSNMPClient)
- [x] TopologyService.Export (json/graphml/dot/text)
- [x] ScannerService.Stop (graceful shutdown)
- [x] AuditService, DeviceControlService, WOL, NetTools
- [x] Plugin system (OSFilter, CSVExporter, dynamic loader)
- [x] ARP cache (асинхронное кэширование)
- [x] Incremental CLI output
- [x] Adaptive port-scanning limits
- [x] Permissions check (cross-OS)

### Этап v2.2: SRP Refactoring + IPv6
- [x] SRP Refactoring: app.go 2780→605 строк (-78%)
- [x] 12 новых модульных файлов в gui/
- [x] IPv6 Support (dual-stack detection, GUI indicators)
- [x] GUI Smoke Tests (24 smoke-теста)

### Этап D-Track: Stabilization
- [x] Topology hardening (confidence downgrade, deterministic sort)
- [x] Export hardening (JSON validation, GraphML keys)
- [x] GUI UX hardening (pagination, presets, analytics)

### Этап v2.3.1: Единый CLI + inventory-подкоманды (2026-09-22)
- [x] Дедупликация диспатча: `ExecuteCLI` делегирует в cobra `rootCmd` (удалён legacy-`switch`)
- [x] `inventory list` — реализация + `--limit` / позиционный limit
- [x] `inventory diff` — реализация + `--history` (comparator/port-changes)
- [x] `inventory save` — реальное сканирование → снапшот (`--hosts-file`/`--network`, `--id`)
- [x] Подкоманда `gui` в cobra root; общий persistent-флаг `--db`
- [x] Тесты `cli_wiring_test.go` (13), coverage `cmd` 0% → 11.6%, lint 0

### Этап M2: Устранение гонок данных (2026-09-24) ✅
- [x] `internal/ports`: устранена гонка в `formatIANAServiceName` — общий `cases.Caser`
      (`golang.org/x/text/cases`) из пакетного состояния вызывался параллельно из
      горутин сканирования (`scanner.scanHost`); заменён чистой `titleCaseWord`
      (без x/text, без `recover`), убран неиспользуемый `go.mod`-импорт
- [x] `internal/network`: `TestARPCacheRapidRefresh` читал счётчик refreshFunc
      без синхронизации — переведён на `atomic`
- [x] Регрессионные тесты: `TestLookupServiceName_Concurrent`,
      `TestTitleCaseWord`, `TestFormatIANAServiceName_Whitespace` (`internal/ports`)
- [x] Проверка: `go test -race ./... -short` — **0 FAIL, 0 DATA RACE**

### Этап E: Аудит и security hardening (2026-09-23)
- [x] E0 Regression `internal/topology` — исправлен (48/48 пакетов green)
- [x] E1 Очистка репозитория: удалена `internal/legacy/` (14 файлов мусора)
- [x] E2 Аудит документации: `docs/audit/01–11` + `AUDIT_REPORT.md`
- [x] E5 Coverage: scanner 84.7%, topology 88.3%, apperror 98.4%, commands 97.4%
- [x] P0-1 Bearer Token auth middleware для REST API (+ 6 тестов)
- [x] P0-2 remote-exec: dry-run **по умолчанию** (`--execute` + `--consent I_UNDERSTAND`)
- [x] P0-3 device-control: reboot требует `--confirm I_UNDERSTAND` (проверка и в сервисе)
- [x] P1-1 `docs/CLI_REFERENCE.md` — полный справочник CLI-флагов
- [x] P1-2 `CONTRIBUTING.md`: trunk-based branching + Conventional Commits + CODEOWNERS
- [x] P1-3 `.github/CODEOWNERS` — карта ответственности
- [x] P1-5 Унификация `CHANGELOG.md` (дубликат `docs/CHANGELOG.md` удалён)
- [x] P1-6 `QUICKSTART_WINDOWS_BUILD.md` перенесён в `docs/`
- [x] P1-7 Объединение `ROADMAP.md` + `IMPLEMENTATION_PLAN.md` в этот документ

## Текущий план

- **Аудит (канонический отчёт):** [../AUDIT_REPORT.md](../AUDIT_REPORT.md)
- **Артефакты аудита:** [audit/README.md](audit/README.md)
- **Единый оптимизированный план:** [UNIFIED_OPTIMIZED_PLAN_2026-09-15.md](UNIFIED_OPTIMIZED_PLAN_2026-09-15.md) — канонический операционный план
- **Финальный план цикла 2026-09-15 (F1–F8):** [FINAL_PLAN_2026-09-15.md](FINAL_PLAN_2026-09-15.md) — F1–F7 выполнены
- **Три плана анализа:** [THREE_PLANS_ANALYSIS_2026-09-15.md](THREE_PLANS_ANALYSIS_2026-09-15.md)
- **План реализации v2.3 (архив):** [archive/2026-09-cycle/IMPLEMENTATION_PLAN_v1.md](archive/2026-09-cycle/IMPLEMENTATION_PLAN_v1.md)

## 🟡 MEDIUM: Текущий приоритет

### M1: Повышение core coverage до 85%+
**Текущее:** scanner 84.7%, api 82.6%, devicecontrol 73.5% | **Цель:** 85%+ | **Оценка:** 4-5 дней

**Что нужно сделать:** дописать тесты для критических пакетов (scanner, api, devicecontrol).

**Критерии завершения:**
- [ ] `go test -cover ./internal/scanner/... ./internal/api/... ./internal/devicecontrol/...` ≥ 85%
- [ ] Нет regressions в существующих тестах

### M2: CI/CD стабилизация
- [x] Lint `golangci-lint run` — 0 замечаний
- [x] Test `go test ./... -short` — 52/52 пакета ok
- [x] `go test -race ./... -short` — 0 FAIL, 0 DATA RACE (2026-09-24)
- [x] Добавить `govulncheck` в CI
- [x] Добавить `docker-compose.yml` + CI-сборка образа

### M3: Улучшение документации
- [x] `docs/CLI_REFERENCE.md`
- [x] `.github/CODEOWNERS`
- [x] `CONTRIBUTING.md` (branching, commits, ответственность)
- [ ] Актуализировать `docs/ARCHITECTURE.md` под v2.3
- [ ] Автогенерация docs (godoc → `docs/dev/`)
- [ ] Обновить `docs/GUI.md`

### M4: Добавить performance benchmarks
- [x] Пакет `internal/benchmark` (замеры сканирования)
- [ ] Бенчмарки для парсинга/экспорта (topology, report)

## 🟢 LOW: По желанию

### L1: UI theme customization
**Текущее:** Default theme | **Цель:** Пользовательские темы | **Оценка:** 1-2 дня
- [ ] L1.1 Light/Dark/System
- [ ] L1.2 Кастомные цвета
- [ ] L1.3 Сохранение настроек темы
- [ ] L1.4 Preview темы

### L2: Plugin system (реальные плагины)
- [x] L2.1 Plugin interface + registry (`internal/plugin`, `internal/scanner/plugin`)
- [ ] L2.2 Загрузка плагинов из директории (dynamic loader)
- [ ] L2.3 Примеры плагинов (filter, export, alert)
- [ ] L2.4 Гайд по созданию плагинов

### L3: Mobile support — ❌ НЕ ПЛАНИРУЕТСЯ (отменено 2026-09-19)
**Решение:** мобильный GUI не поставляется.

> Задачи L3.1–L3.4 ранее ошибочно отмечались выполненными в
> `PLAN_STABILIZATION_V2.md` (заглушки с TODO были приняты за готовность).
> Фактически `gomobile bind` собирает только Go-библиотеку, а не Fyne-GUI;
> мёртвый код (`mobile_layout.go`, `touch_gestures.go`) и нерабочие скрипты
> (`build-android.*`, `build-ios.sh`) удалены. Если мобильный GUI потребуется —
> оформить отдельным эпиком (~3–5 недель: FyneApp.toml, `fyne package`,
> CI, touch-жесты, тестирование на устройствах).

### L4: Telemetry (опционально)
- [ ] L4.1 Опциональная анонимная телеметрия
- [ ] L4.2 Opt-out + privacy policy
- [ ] L4.3 Error reporting с consent
- [ ] L4.4 Документация телеметрии

## 🔴 SECURITY: Остаточные риски (из аудита)

- [x] REST API: Bearer Token auth (P0-1)
- [x] remote-exec: dry-run по умолчанию (P0-2)
- [x] device-control: confirm для reboot (P0-3)
- [ ] TLS в remoteexec — best-effort, требуется строгий режим верификации
- [ ] Аудит-лог для всех изменяющих операций (частично реализован)

## 📅 План релизов

### v2.3.x (Q3 2026) — текущая серия
**Входящие задачи:**
- M1 (core coverage 85%+)
- M2 (govulncheck, docker-compose)
- M3 (ARCHITECTURE.md, автогенерация docs, GUI.md)

**Критерии выхода:**
- [ ] Coverage core ≥ 85%
- [ ] Все CI checks проходят (lint, test, build, govulncheck)
- [ ] Документация обновлена
- [ ] Нет critical bugs

### v2.4.0 (Q4 2026)
**Цель:** Интеграция архитектурного слоя v2.3
- E6: eventbus в scan-цикл
- E6: configvalidation в CLI startup
- Event-driven GUI через eventbus
- Расширения E7 (ICMP, UDP, PDF, plugin-probes)

### v3.0.0 (Q1 2027)
- Метрики Prometheus + structured logging
- Мультиплатформенное развёртывание (Docker, systemd)

## 📋 Метрики качества

### Code Coverage (2026-09-23)
| Пакет | Coverage | Цель |
|-------|----------|------|
| scanner | 84.7% | 85% |
| topology | 88.3% | ≥80% |
| apperror | 98.4% | ≥90% |
| commands | 97.4% | ≥90% |
| api | 82.6% | 85% |
| devicecontrol | 73.5% | 85% |

### Performance
- CLI размер: ~60.6 MB
- GUI размер: ~58.5 MB
- Планируется: ldflags `-s -w` + upx для уменьшения

### Code Quality
- `go build ./...` — чисто
- `golangci-lint run` — 0 замечаний
- `go test ./... -short` — 48/48 ok

## 🔄 Процесс разработки

Ветвление и коммиты — см. [../CONTRIBUTING.md](../CONTRIBUTING.md) (trunk-based, Conventional Commits).

### 1. Planning
Еженедельный обзор roadmap; новые задачи — только в этот документ.

### 2. Development
Короткие PR в `main` (trunk-based), Conventional Commits.

### 3. Review
Обязательный review по [CODEOWNERS](../.github/CODEOWNERS), CI green (lint, test, build).

### 4. Release
Тег `vX.Y.Z`, обновление [../CHANGELOG.md](../CHANGELOG.md), артефакты в `build/`.

## 📝 Notes

### Зависимости между задачами
- M1 (coverage) не зависит от E6 (интеграция слоя).
- E6 (eventbus/configvalidation) требует стабильного core (M1).

### Риски
- Рост связанности при интеграции eventbus → mitigation: событийный контракт в `internal/contracts`.
- Coverage GUI остаётся низким → приоритет на бизнес-логику (controller), а не UI-инициализацию.

## Ссылки

- Структура проекта: [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md)
- Архитектура: [ARCHITECTURE.md](ARCHITECTURE.md)
- Техническая документация: [TECHNICAL.md](TECHNICAL.md)
- Руководство пользователя: [USER_GUIDE.md](USER_GUIDE.md)
- CLI-справочник: [CLI_REFERENCE.md](CLI_REFERENCE.md)
- История изменений: [../CHANGELOG.md](../CHANGELOG.md)
- Архив завершённых документов: [archive/](archive/)

## Maintenance rules

- Обновлять этот файл при изменениях структуры roadmap и плана.
- Единственный актуальный операционный план — `UNIFIED_OPTIMIZED_PLAN_2026-09-15.md`; новые датированные этапы добавлять только туда.
- Завершённые планы и отчёты перемещать в `docs/archive/<YYYY-MM-cycle>/` и обновлять ссылки.
- Не создавать дублирующих индексных документов — индекс навигации ведётся в [README.md](README.md).
