# Audit Report: Documentation vs Code

**Дата аудита:** 2026-09-02  
**Аудитор:** Koda AI  
**Цель:** Выявить несоответствия документации коду, устаревшие файлы, области для оптимизации

---

## 1. КРИТИЧЕСКИЕ НЕСООТВЕТСТВИЯ (BLOCKER)

### 1.1 Архитектура GUI — ДОКУМЕНТАЦИЯ НЕ СООТВЕТСТВУЕТ КОДУ

**ARCHITECTURE.md (строки 74-79)** описывает монолитную структуру:
```
internal/gui/
├── app.go          # Главный файл GUI
├── scan_controller.go    # Контроллер сканирования
├── topology_controller.go # Контроллер топологии
├── results_view.go       # Отображение результатов
└── operations.go         # Operations Runtime
```

**ФАКТИЧЕСКАЯ СТРУКТУРА (108 файлов в internal/gui/):**
- `app.go` — 605 строк (был 2780, рефакторинг -78%)
- `init_ui.go` — инициализация UI (новое)
- `scan_observer.go` — наблюдение за раннером (новое)
- `results_pipeline.go` — фильтрация, сортировка, кэш
- `results_table_view.go` — таблица, карточки
- `results_filters_ui.go` — filter presets
- `results_settings_ui.go` — сохранение настроек
- `results_analytics_view.go` — аналитика
- `operations_ui.go` — operations center
- `tools_ui.go` — ping, traceroute, DNS, WOL
- `event_handlers.go` — обработчики событий
- `ui_helpers.go` — badge, theme, menu
- `controller/` — 44 файла контроллеров
- `errors/` — 3 файла ошибок

**ВЫВОД:** Архитектура полностью переработана (SRP refactoring), документация не обновлена.

---

### 1.2 IPv6 Support — ДОКУМЕНТАЦИЯ НЕ УПОМИНАЕТ

**ФАКТ:** В коде добавлена полная поддержка IPv6:
- `internal/network/ipv6.go` — 100 строк (IsIPv4, IsIPv6, IsIP, IsCIDR, FormatIPForDisplay)
- `internal/network/network.go` — обновлён DetectLocalNetwork для IPv6 fallback
- `internal/gui/results_table_view.go` — форматирование `[IPv6] (v6)`
- `internal/network/ipv6_test.go` — 270 строк тестов

**ДОКУМЕНТАЦИЯ:**
- `ARCHITECTURE.md` — упоминает только IPv4
- `ROADMAP.md` — не упоминает IPv6
- `TECHNICAL.md` — нет IPv6

**ВЫВОД:** IPv6 feature не задокументирован.

---

### 1.3 Coverage Metrics — УСТАРЕВШИЕ ДАННЫЕ

**ROADMAP.md (строки 17-28)** заявляет:
```
internal/gui           | 51.6% | ✅ Стабильно
internal/gui/controller| 44.4% | ✅ Стабильно
```

**ФАКТИЧЕСКИЕ ДАННЫЕ (после рефакторинга):**
- GUI coverage изменился из-за разделения 10 файлов
- Smoke tests добавлены: `gui_smoke_test.go` (24 теста)
- IPv6 tests добавлены: `ipv6_test.go` (10 тестов)

**ВЫВОД:** Metrics требуют актуализации после рефакторинга.

---

## 2. УСТАРЕВШИЕ ССЫЛКИ В ДОКУМЕНТАЦИИ

### 2.1 IMPLEMENTATION_PLAN.md

| Строка | Устаревшее | Факт |
|--------|-----------|------|
| 40 | "GUI интерфейс (Fyne framework, 30+ файлов)" | 108 файлов в gui/ |
| 16-17 | Coverage GUI 51.6% | Требуется пересчёт |
| 84 | M1.1 Тесты для scanner/UDP | Уже добавлены в scanner_service.go |
| 105 | M2.1 CI/CD pipeline | Не реализован |

### 2.2 ARCHITECTURE.md

| Строка | Устаревшее | Факт |
|--------|-----------|------|
| 74-79 | Монолитная структура gui/ | 108 файлов, SRP разделение |
| 183-184 | "internal/gui/app.go" | app.go теперь 605 строк (был 2780) |
| 194-199 | "internal/gui/controllers" | controller/ — 44 файла |
| 603 | "Последнее обновление: 2026-01-XX" | Должно быть 2026-09-02 |

### 2.3 ROADMAP.md

| Строка | Устаревшее | Факт |
|--------|-----------|------|
| 8-10 | "v2.2 Backlog: ✅ 13/13" | Задачи P1-P4 завершены, но не обновлены |
| 24 | scanner 71.9% | Требуется пересчёт |
| 26 | network 85.3% | IPv6 tests добавлены |

---

## 3. АРХИВНЫЕ ФАЙЛЫ — ПЕРЕНОС В АРХИВ

### 3.1 Критически устаревшие (перенести):

| Файл | Причина | Дата создания |
|------|---------|---------------|
| `docs/INSTALL_OLD.md` | Заменён INSTALL.md | 2026-01 |
| `docs/DETAILED_BACKLOG_P3_STAGE2.md` | Backlog v21/v22 актуален | 2026-01 |
| `docs/ROADMAP_P1_P3.md` | Заменён ROADMAP.md | 2026-01 |
| `docs/TASK_BACKLOG_V21.md` | Backlog v22 актуален | 2026-01 |
| `docs/RELEASE_NOTES_V22.md` | Актуальны RELEASE_NOTES_1.0.x | 2026-01 |
| `docs/DETAILED_PLAN_V22.md` | Заменён IMPLEMENTATION_PLAN.md | 2026-01 |
| `docs/RELEASE_MANIFEST_V2.md` | Неактуален | 2026-01 |
| `docs/RELEASE_INSTRUCTIONS_V22.md` | Неактуален | 2026-01 |
| `docs/PUBLISH_RELEASE_V22.md` | Неактуален | 2026-01 |
| `docs/BLOCKER_ZACILIVANIE_001.md` | Blocker закрыт | 2026-01 |
| `docs/RELEASE_ACCEPTANCE_CHECKLIST.md` | Неактуален | 2026-01 |
| `docs/RELEASE_READINESS_SNAPSHOT.md` | Неактуален | 2026-01 |
| `docs/CHECKLIST_STATUS_INDEX.md` | Неактуален | 2026-01 |
| `docs/P0_SIGNOFF_RUNBOOK.md` | Неактуален | 2026-01 |
| `docs/STAGE2_100_COMMIT_READY.md` | Неактуален | 2026-01 |
| `docs/RELEASE_OPERATIONS_CHEATSHEET.md` | Неактуален | 2026-01 |
| `docs/DOCUMENTATION_UPDATE_REPORT_2026-08-18.md` | Аудит завершён | 2026-08 |

### 3.2 Архивные папки — проверка:

| Папка | Содержимое | Действие |
|-------|-----------|----------|
| `docs/archive/2026-01-release-cycle/` | 31 файл | ✅ Оставить (история) |
| `docs/archive/2026-04-docs-sync/` | 5 файлов | ✅ Оставить |
| `docs/archive/2026-08-ui-tests/` | 3 файла | ✅ Оставить |
| `docs/archive/2026-09-audit/` | НОВАЯ | ✅ Создать для отчётов |

---

## 4. АНАЛИЗ ОПТИМИЗАЦИИ КОДА

### 4.1 Дублирование кода

| Пакет | Дублирование | Оценка |
|-------|-------------|--------|
| `internal/gui/` | 44 test-файла, ~5000 строк тестов | ✅ OK (good practice) |
| `internal/network/` | `parseIPv4NetworkRange` и `parseIPv6NetworkRange` | ✅ OK (разная логика) |
| `internal/scanner/` | `isHostAlive` с 3 методами | ✅ OK (strategy pattern) |

### 4.2 Устаревшие import'ы

| Файл | Устаревший import | Замена |
|------|------------------|--------|
| `internal/gui/app.go` | `scand "network-scanner/internal/scanner/daemon"` | ✅ Актуален |

### 4.3潜在的的性能问题

| Метод | Проблема | Рекомендация |
|-------|----------|-------------|
| `buildTableView` | 8 колонок, пересоздание на каждый рендер | Кэшировать cells |
| `buildCardsView` | widget.NewList без pooling | Использовать fyne v2.5+ ListPool |
| `scheduleResultsRender` | debounce 180ms | Оптимизировать под hardware |

### 4.4 Code Smells

| Место | Проблема | Severity |
|-------|----------|----------|
| `app.go:556-841` (было) | initUI был 285 строк | ✅ Исправлено (init_ui.go) |
| `results_view.go` | 775 строк | 🟡 Подразделить |
| `controller/scan_controller.go` | 420 строк | 🟡 Подразделить |

---

## 5. ПЛАН УСТРАНЕНИЯ НЕСООТВЕТСТВИЙ

### PRIORITY 1: КРИТИЧЕСКИЕ (сделать сейчас)

| # | Задача | Файл | Время | Статус |
|---|--------|------|-------|--------|
| 1.1 | Обновить ARCHITECTURE.md | docs/ARCHITECTURE.md | 2 часа | ⬜ |
| 1.2 | Обновить ROADMAP.md | docs/ROADMAP.md | 1 час | ⬜ |
| 1.3 | Обновить IMPLEMENTATION_PLAN.md | docs/IMPLEMENTATION_PLAN.md | 1 час | ⬜ |
| 1.4 | Добавить IPv6 в DOCUMENTATION | docs/TECHNICAL.md | 1 час | ⬜ |

### PRIORITY 2: АРХИВИРОВАНИЕ (сделать сейчас)

| # | Задача | Файлы | Время | Статус |
|---|--------|-------|-------|--------|
| 2.1 | Перенести устаревшие docs | 17 файлов | 30 мин | ⬜ |
| 2.2 | Создать ARCHIVE_INDEX.md | docs/archive/ | 15 мин | ⬜ |

### PRIORITY 3: ОПТИМИЗАЦИЯ (спланировать)

| # | Задача | Оценка | Статус |
|---|--------|--------|--------|
| 3.1 | Подразделить results_view.go | 2 часа | ⬜ |
| 3.2 | Подразделить scan_controller.go | 2 часа | ⬜ |
| 3.3 | Добавить ListPool для cards | 1 час | ⬜ |
| 3.4 | Кэшировать table cells | 1.5 часа | ⬜ |

---

## 6. ИТОГОВАЯ СТАТИСТИКА

| Метрика | Значение |
|---------|----------|
| Всего файлов в docs/ | 104 |
| Устаревших файлов | 17 |
| Требуют обновления | 4 |
| Актуальных файлов | 83 |
| Файлов в internal/gui/ | 108 |
| Строк в internal/gui/ | ~8500 |
| Тестов в internal/gui/ | ~500 |
| Coverage GUI (оценка) | ~55% |
| IPv6 Support | ✅ Реализован |
| SRP Refactoring | ✅ Завершён |

---

**Отчёт подготовлен:** 2026-09-02  
**Следующий аудит:** 2026-10-02  
**Статус:** 🔴 КРИТИЧЕСКИЕ несоответствия требуют устранения
