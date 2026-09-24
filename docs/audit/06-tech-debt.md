# Технический долг

**Дата:** 2026-09-22  

---

## 6.1. Regression в `internal/topology`

| Аспект | Статус |
|--------|--------|
| Текущий статус | ✅ **Исправлен** (F1/F2, 2026-09-15, 48/48 пакетов green) |
| Причина (была) | `topologyServiceImpl.Export` — заглушка TODO; `TestSaveGraphMLToBytes` — фикстура вне `Validate()` |
| Решение | Реализация `Export` (6 форматов); добавлен `sortedDeviceKeys`; дедупликация `topology_handlers.go` |
| Критерии | `go test ./...` — 49/49 пакетов green; `Export` пишет файл |

**Оставшиеся риски:**
- Возможна проблема в CI или в интеграционных сценариях (GUI → topology → inventory)
- Рекомендовано: добавить `go test -race` в CI; добавить интеграционные тесты topology → inventory

---

## 6.2. Coverage-разрывы

| Пакет | Текущий | Цель | Статус |
|-------|---------|------|--------|
| `internal/scanner` | 84.7% | 85% | ⚠️ Близко |
| `internal/api` | 82.6% | 85% | ⚠️ |
| `internal/devicecontrol` | 73.5% | 85% | ❌ |
| `internal/topology` | 88.3% | 85% | ✅ |

---

## 6.3. Архитектурный слой v2.3

| Модуль | Статус | Интеграция |
|--------|--------|-----------|
| `internal/apperror` | ✅ 98.4% | Самодостаточная |
| `internal/commands` | ✅ 97.4% | Самодостаточная |
| `internal/eventbus` | ✅ 93.5% | Самодостаточная |
| `internal/plugin` | ✅ ~60% | Самодостаточная |
| `internal/configvalidation` | ✅ | Самодостаточная |
| `internal/benchmark` | ✅ | Не используется в CI |

**Решение (E6):** Инкрементальная обвязка — не ретрофитировать в легаси.

---

## 6.4. Артефакты покрытия

| Путь | Описание | Статус |
|------|----------|--------|
| `internal/legacy/` (14 файлов) | `api_cov.out`, `cov.out`, `coverage_*.out`, `GUI_STRUCTURE.txt`, `launch-gui.sh`, `LOG-*.txt`, `README.md`, `topo_cov.out` | ⚠️ Удалить / перенести в архив |
| Корень (cov-файлы) | `api_cov`, `banner_cov`, `scanner_cov*`, `*_test*.log` | ✅ Удалены (F4) |
| `.gitignore` | Паттерны cov-артефактов, `*.log`, `.tmp/` | ✅ Дополнен (F5) |

---

## 6.5. Cobra-CLI

| Аспект | Статус |
|--------|--------|
| `spf13/cobra` | indirect dependency |
| `cmd/network-scanner/cmd/scan_cobra.go` | Используется реальная бизнес-логика |
| Интеграция | ✅ Активна: `main → ExecuteCLI → scanCmd.RunE` |

---

## 6.6. TODO / FIXME / HACK

- **Remote-exec без dry-run по умолчанию** — `[RISK]`
- **Device-control без confirm по умолчанию** — `[RISK]`
- **Security report без redaction по умолчанию** — `[RISK]`
- **TLS best-effort в remoteexec** — `[RISK]`
- **API без аутентификации** — `[RISK]`
- **Планировщик периодических сканов** — идея (Plan 3.6)
- **PDF-отчёты** — идея (Plan 3.6)

---

## 6.7. `internal/legacy/`

| Файл | Размер | Описание | Рекомендация |
|------|--------|----------|--------------|
| `api_cov.out` | 24 KB | Coverage-артрафт | Удалить |
| `cov.out` | 31 KB | Coverage-артрафт | Удалить |
| `coverage_ctrl.out` | 52 KB | Coverage-артрафт | Удалить |
| `coverage_gui.out` | 133 KB | Coverage-артрафт | Удалить |
| `coverage_scanner.out` | 31 KB | Coverage-артрафт | Удалить |
| `gui_cov.out` | 141 KB | Coverage-артрафт | Удалить |
| `GUI_STRUCTURE.txt` | 25 KB | Структура GUI | Перенести в docs |
| `launch-gui.sh` | 2.5 KB | Устаревший скрипт | Архив (GUI.md пометил как [ARCHIVED]) |
| `LOG-network-scanner-gui-1.0.3.txt` | 43 KB | Лог | Удалить |
| `README.md` | 1.8 KB | Legacy README | Перенести/удалить |
| `topo_cov.out` | 30 KB | Coverage-артрафт | Удалить |
| `build-release-windows-only.ps1` | 3.8 KB | Legacy build | Архив |

---