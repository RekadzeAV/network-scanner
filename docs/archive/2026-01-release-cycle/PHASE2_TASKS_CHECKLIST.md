# Phase 2 Tasks Checklist

## 2.1 Live Inventory

- [x] Создать пакет `internal/inventory`
  - [x] SQLite storage (`snapshots`)
  - [x] `SaveSnapshot(scanID, timestamp, []Host)`
  - [x] `LoadSnapshot(scanID)`
  - [x] `Diff(scanIDA, scanIDB)`
  - [x] Unit tests для save/load/diff
- [x] CLI флаги и интеграция
  - [x] `--inventory-save`
  - [x] `--inventory-diff`
  - [x] `--inventory-db`
- [ ] GUI inventory tab
  - [x] Выбор snapshot A/B
  - [x] Визуализация new/missing/changed
  - [x] Автосохранение snapshot после успешного сканирования (опция GUI)

## 2.2 Device Profiling

- [x] Создать/выделить `deviceclassifier`
- [x] Ввести фиксированный список категорий:
  - [x] Unknown
  - [x] Router/Switch
  - [x] Access Point
  - [x] Printer
  - [x] Camera
  - [x] NAS
  - [x] IoT
  - [x] Desktop/Laptop
  - [x] Server
  - [x] Phone/Tablet
- [~] Обновить эвристики (OUI + ports + banners + SNMP SysDescr)
  - [x] Port-pattern heuristics (baseline)
  - [ ] OUI enrichment in classifier layer
  - [ ] Banner-aware heuristics
  - [ ] SNMP SysDescr-aware heuristics
- [ ] Добавить иконки категорий в GUI
- [~] Добавить иконки категорий в GUI
  - [x] Бейджи категорий в таблице и карточках (`[NET]`, `[SRV]`, ...)
  - [ ] Перейти на графические Fyne-иконки
- [ ] Использовать категории для цветов топологии
- [x] Unit tests для classifier

## 2.3 Human-Centric Security

- [x] `audit.HumanReadable(finding)`
- [x] Индекс безопасности 0..100
  - [x] critical=30, high=20, medium=10, low=5
  - [x] clamp в диапазон [0..100]
- [x] Отображение индекса в Security Dashboard (базовый индикатор)
- [x] Использование human-readable текста в dashboard таблице для port-audit findings
- [ ] Полный UX-полиш Security Dashboard для нетехнического режима
  - [ ] визуальный color-widget индикатор
  - [ ] компактный onboarding текст
