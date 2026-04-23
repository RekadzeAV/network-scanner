# Network Scanner - Сканер локальной сети

[![GitHub](https://img.shields.io/github/license/RekadzeAV/network-scanner)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)

Кроссплатформенная утилита для сканирования локальных сетей с детальной аналитикой.

**Репозиторий:** https://github.com/RekadzeAV/network-scanner

## 🚀 Быстрый старт

### CLI версия (командная строка)

```bash
# Сборка
go build -o network-scanner ./cmd/network-scanner

# Запуск
./network-scanner
```

### GUI версия (графический интерфейс)

```bash
# Сборка
go build -o network-scanner-gui ./cmd/gui

# Запуск
./network-scanner-gui
```

### Что нового в GUI

- Подрежимы `Devices/Security` на вкладке сканирования.
- `Host Details Drawer` с быстрыми действиями (`Ping/Traceroute/DNS/Whois/WOL`).
- `Operations Center` в `Инструменты` с историей и действиями `Retry/Cancel`.
- `Security Dashboard` с агрегированными findings и экспортом HTML-отчета.

## 📁 Структура проекта

```
Сканер локальной сети/
├── cmd/
│   ├── network-scanner/    # Точка входа CLI приложения
│   └── gui/                # Точка входа GUI приложения
├── internal/
│   ├── scanner/           # Логика сканирования
│   ├── network/           # Работа с сетью
│   ├── display/           # Отображение результатов (CLI)
│   └── gui/               # Компоненты графического интерфейса
├── docs/                  # Документация
├── scripts/               # Скрипты сборки
└── README.md             # Этот файл
```

## 📚 Документация

Полная документация находится в папке [docs/](docs/):

- **[Инструкция по эксплуатации](Инструкция%20по%20эксплуатации.md)** - Полная инструкция по эксплуатации программы (русский язык)
- **[README.md](docs/README.md)** - Основная документация с описанием возможностей
- **[USER_GUIDE.md](docs/USER_GUIDE.md)** - Подробное руководство пользователя
- **[GUI.md](docs/GUI.md)** - Документация по GUI версии приложения
- **[INSTALL.md](docs/INSTALL.md)** - Инструкции по установке
- **[QUICKSTART-macOS.md](docs/QUICKSTART-macOS.md)** - Быстрый старт для macOS
- **[TECHNICAL.md](docs/TECHNICAL.md)** - Техническая документация
- **[ARCHITECTURE.md](docs/ARCHITECTURE.md)** - Архитектура проекта
- **[ROADMAP.md](docs/ROADMAP.md)** - Канонический роадмап проекта
- **[TEST_NETWORK_PROFILE.md](docs/TEST_NETWORK_PROFILE.md)** - Фиксированный профиль тестовой сети (подсеть, шлюз, DNS, NAS)
- **[P1_CLOSURE_CHECKLIST.md](docs/P1_CLOSURE_CHECKLIST.md)** - Чеклист формального закрытия этапа P1
- **[ANALYSIS.md](docs/ANALYSIS.md)** - Анализ проекта
- **[RELEASE_SUMMARY_UI_RESULTS.md](docs/RELEASE_SUMMARY_UI_RESULTS.md)** - Краткий релиз-итог по UI результатов
- **[RELEASE_SUMMARY_STAGE2_P2.md](docs/RELEASE_SUMMARY_STAGE2_P2.md)** - Краткий релиз-итог по Stage2 P2 (Risk Signatures + Device Control)
- **[RELEASE_SUMMARY_STAGE2_P3.md](docs/RELEASE_SUMMARY_STAGE2_P3.md)** - Краткий релиз-итог по Stage2 P3 (CVE + Security Report + Remote Exec)
- **[PR_DESCRIPTION_UI_RESULTS.md](docs/PR_DESCRIPTION_UI_RESULTS.md)** - Готовое описание PR по доработкам UI результатов
- **[UI_IMPLEMENTATION_BACKLOG.md](docs/UI_IMPLEMENTATION_BACKLOG.md)** - Актуализированный backlog UI-рефакторинга (`P0..P5`)
- **[RELEASE_ACCEPTANCE_CHECKLIST.md](docs/RELEASE_ACCEPTANCE_CHECKLIST.md)** - Финальный чеклист приемки перед релизом
- **[P3_CLOSURE_CHECKLIST.md](docs/P3_CLOSURE_CHECKLIST.md)** - Формальное закрытие Stage 1 / P3 и CI sign-off
- **[RELEASE_OPERATIONS_CHEATSHEET.md](docs/RELEASE_OPERATIONS_CHEATSHEET.md)** - Краткий набор команд для closure-прогонов и релизного дежурства
- **[BUILD_STRUCTURE.md](docs/BUILD_STRUCTURE.md)** - Структура каталогов релизных скриптов (`build/release/`)
- **[RELEASE_READINESS_SNAPSHOT.md](docs/RELEASE_READINESS_SNAPSHOT.md)** - Текущий срез готовности релиза (авто/ручные шаги)
- **[CHECKLIST_STATUS_INDEX.md](docs/CHECKLIST_STATUS_INDEX.md)** - Единый индекс статусов по всем checklist-документам
- **[RELEASE_READY_GAP_LIST.md](docs/RELEASE_READY_GAP_LIST.md)** - Приоритизированный backlog оставшихся шагов до финального sign-off
- **[P0_SIGNOFF_RUNBOOK.md](docs/P0_SIGNOFF_RUNBOOK.md)** - Пошаговый runbook закрытия блокирующего P0 (Cross-OS + CI evidence)
- **[STAGE2_100_COMMIT_READY.md](docs/STAGE2_100_COMMIT_READY.md)** - Краткий commit-ready итог по Stage2 и remaining шагам sign-off
- **[RELEASE_READINESS_PR_SNIPPET.md](docs/RELEASE_READINESS_PR_SNIPPET.md)** - Готовый блок статуса для PR/релиз-комментария
- **[RELEASE_READINESS_STAGE2_P3_PR_SNIPPET.md](docs/RELEASE_READINESS_STAGE2_P3_PR_SNIPPET.md)** - Готовый PR-блок для Stage2/P3 (EN/RU, short/long)
- **[RELEASE_READINESS_PR_READY.md](docs/RELEASE_READINESS_PR_READY.md)** - Короткий и расширенный ready-to-paste блок для PR
- **[DOCS_SYNC_SUMMARY_2026-04-23.md](docs/DOCS_SYNC_SUMMARY_2026-04-23.md)** - Сводка синхронизации документации (флаги/Go baseline)
- **[DOCS_SYNC_PR_SNIPPET_2026-04-23.md](docs/DOCS_SYNC_PR_SNIPPET_2026-04-23.md)** - Короткий ready-to-paste RU блок для PR
- **[DOCS_SYNC_PR_SNIPPET_2026-04-23_EN.md](docs/DOCS_SYNC_PR_SNIPPET_2026-04-23_EN.md)** - Короткий ready-to-paste EN блок для PR
- **[FINAL_PR_COMMENT_READY.md](docs/FINAL_PR_COMMENT_READY.md)** - Финальный ready-to-paste комментарий для PR
- **[FINAL_PR_COMMENT_STAGE2_P3_READY.md](docs/FINAL_PR_COMMENT_STAGE2_P3_READY.md)** - Финальный ready-to-paste комментарий для PR по Stage2/P3 (RU)
- **[MANUAL_SIGNOFF_TEMPLATE.md](docs/MANUAL_SIGNOFF_TEMPLATE.md)** - Шаблон ручного sign-off перед релизом
- **[MANUAL_SIGNOFF_DRAFT.md](docs/MANUAL_SIGNOFF_DRAFT.md)** - Черновик sign-off с предзаполненными авто-evidence
- **[GRAPHML_COMPATIBILITY_CHECK.md](docs/GRAPHML_COMPATIBILITY_CHECK.md)** - Ручная проверка совместимости GraphML (yEd/Gephi)
- **[D_TRACK_EVIDENCE_TEMPLATE.md](docs/D_TRACK_EVIDENCE_TEMPLATE.md)** - Шаблон evidence-блока для D-track hardening
- **[D_TRACK_EVIDENCE_CURRENT.md](docs/D_TRACK_EVIDENCE_CURRENT.md)** - Текущий снимок статуса D-track evidence для ветки
- **[D_TRACK_EVIDENCE_PR_SNIPPET.md](docs/D_TRACK_EVIDENCE_PR_SNIPPET.md)** - Короткий ready-to-paste блок для описания PR
- **[DEVELOPMENT_MAP.md](DEVELOPMENT_MAP.md)** - Детальная карта разработки проекта
- **[ROADMAP_P1_P3.md](docs/ROADMAP_P1_P3.md)** - Дорожная карта приоритетов P1–P3 (два этапа развития)
- **[DETAILED_BACKLOG_P3_STAGE2.md](docs/DETAILED_BACKLOG_P3_STAGE2.md)** - Детализированный backlog задач по Этапу 1 P3 и Этапу 2 P1/P2/P3
- **[CHANGELOG.md](CHANGELOG.md)** - История изменений проекта
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Правила вклада и соглашения по коммитам
- **[QUICKSTART_WINDOWS_BUILD.md](QUICKSTART_WINDOWS_BUILD.md)** - Быстрый старт: сборка для Windows на macOS
- **[RELEASE_NOTES_1.0.3.md](RELEASE_NOTES_1.0.3.md)** - Примечания к релизу 1.0.3

## 🔧 Сборка

### Единые точки входа (Makefile)

```bash
make build
make test
make run
make deploy
```

### Первичная настройка окружения (bootstrap)

```bash
# Linux/macOS
./scripts/bootstrap.sh
```

```powershell
# Windows PowerShell
.\scripts\bootstrap.ps1
```

### Локальная интеграционная проверка

```bash
# Linux/macOS
./scripts/integration-check.sh

# или через go test c integration-тегом
go test -tags=integration ./...

# либо через Makefile
make test-integration
```

```powershell
# Windows PowerShell
.\scripts\integration-check.ps1

# или напрямую:
go test -tags=integration ./...
```

### CLI версия

```bash
go build -o network-scanner ./cmd/network-scanner
```

### GUI версия

```bash
go build -o network-scanner-gui ./cmd/gui
```

### Использование скриптов

```bash
# macOS
./scripts/build-macos.sh

# Linux/Unix
./scripts/build.sh

# Windows (на Windows)
scripts\build.bat

# Сборка для Windows на macOS/Linux (кросскомпиляция)
./scripts/build-windows.sh  # Требует mingw-w64
```

Релизные артефакты этих скриптов (включая `build-release-windows-only.ps1`) создаются в каталоге `build/release/` в корне репозитория, в подпапках вида `YYYY-MM-DD-N/` (например `windows/` для Windows).

### Smoke-проверки (регрессии CLI)

```bash
# Linux/macOS: базовый режим без топологии
./scripts/smoke-cli-no-topology.sh

# Linux/macOS: режим с топологией (проверка SNMP summary)
./scripts/smoke-cli-topology.sh

# Linux/macOS: режим инструментов CLI (ping/dns + raw)
./scripts/smoke-cli-tools.sh
```

```powershell
# Windows PowerShell
.\scripts\smoke-cli-no-topology.ps1
.\scripts\smoke-cli-topology.ps1
.\scripts\smoke-cli-tools.ps1
```

```bash
# Linux/macOS: D-track hardening smoke (topology export consistency)
./scripts/smoke-d-track-topology-export.sh
```

```powershell
# Windows PowerShell
.\scripts\smoke-d-track-topology-export.ps1
```

Оба smoke-скрипта используют `127.0.0.1/32` и короткий диапазон портов для быстрого прогона.
Инструментальный smoke-скрипт использует `127.0.0.1` и `localhost`, а также проверяет вывод секций `raw`.
D-track smoke дополнительно проверяет эквивалентность `json`/`graphml` и fallback-режим экспорта `png` при отсутствии Graphviz.

### Golden-тесты форматированного вывода

```bash
# Запуск golden-тестов (обычный режим)
go test ./internal/display -run Golden

# Обновление golden-снимка при намеренном изменении формата
UPDATE_GOLDEN=1 go test ./internal/display -run Golden
```

```powershell
# Windows PowerShell
go test ./internal/display -run Golden
$env:UPDATE_GOLDEN='1'; go test ./internal/display -run Golden
```

### Закрыть P1 одной командой

```bash
# Linux/macOS (6-command runbook, copy/paste)
./scripts/p1-closure-check.sh && ./scripts/p2-closure-check.sh && ./scripts/p3-closure-check.sh && ./scripts/stage2-p1-closure-check.sh && ./scripts/stage2-p2-closure-check.sh && ./scripts/stage2-p3-closure-check.sh
```

```powershell
# Windows PowerShell (6-command runbook, copy/paste)
.\scripts\p1-closure-check.ps1; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }; .\scripts\p2-closure-check.ps1; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }; .\scripts\p3-closure-check.ps1; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }; .\scripts\stage2-p1-closure-check.ps1; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }; .\scripts\stage2-p2-closure-check.ps1; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }; .\scripts\stage2-p3-closure-check.ps1
```

Подробный операционный вариант и эквиваленты через `make`: `docs/RELEASE_OPERATIONS_CHEATSHEET.md`.
Термины и подписи runbook синхронизированы с `docs/USER_GUIDE.md` и `docs/TECHNICAL.md`.

Проверка локальных markdown-ссылок перед коммитом (Windows):

```powershell
.\scripts\docs-link-check.ps1
# или
make docs-link-check-win
```

```bash
# Linux/macOS
./scripts/p1-closure-check.sh

# либо через Makefile
make p1-check
```

```powershell
# Windows PowerShell
.\scripts\p1-closure-check.ps1

# либо через Makefile (если есть make)
make p1-check-win
```

Сценарий `p1-closure-check` включает: `go test ./...`, smoke-проверки базового CLI, topology-режима и tool-режимов (`ping/dns`).

### Закрыть P2 одной командой

```bash
# Linux/macOS
./scripts/p2-closure-check.sh

# либо через Makefile
make p2-check
```

```powershell
# Windows PowerShell
.\scripts\p2-closure-check.ps1

# либо через Makefile (если есть make)
make p2-check-win
```

Сценарий `p2-closure-check` включает: `go test ./...`, smoke-проверки (`no-topology`, `topology`, `tools`) и целевой sanity для P2-флагов (`--grab-banners`, `--show-raw-banners`, `--os-detect-active`, negative-case для `--wol-mac`).

### Закрыть P3 одной командой

```bash
# Linux/macOS
./scripts/p3-closure-check.sh

# либо через Makefile
make p3-check
```

```powershell
# Windows PowerShell
.\scripts\p3-closure-check.ps1

# либо через Makefile (если есть make)
make p3-check-win
```

Сценарий `p3-closure-check` включает: unit-тесты, integration-тесты (`-tags=integration`), golden-check, perf benchmark и baseline smoke через `p2-closure-check`.

### Закрыть Stage2/P1 одной командой

```bash
# Linux/macOS
./scripts/stage2-p1-closure-check.sh

# либо через Makefile
make stage2-p1-check
```

```powershell
# Windows PowerShell
.\scripts\stage2-p1-closure-check.ps1

# либо через Makefile (если есть make)
make stage2-p1-check-win
```

Сценарий `stage2-p1-closure-check` включает: таргетированные unit-тесты (`internal/nettools`, `internal/audit`, `internal/gui`), smoke tools-check и sanity по `--audit-min-severity` (включая fallback при невалидном значении).

### Закрыть Stage2/P2 одной командой

```bash
# Linux/macOS
./scripts/stage2-p2-closure-check.sh

# либо через Makefile
make stage2-p2-check
```

```powershell
# Windows PowerShell
.\scripts\stage2-p2-closure-check.ps1

# либо через Makefile (если есть make)
make stage2-p2-check-win
```

Сценарий `stage2-p2-closure-check` включает: таргетированные unit-тесты (`internal/risksignature`, `internal/devicecontrol`, `internal/report`, `internal/gui`), smoke tools-check, sanity генерации `security report` (проверка секций `CVE Findings` + `Risk Signature Findings`) и negative-case проверки для `device-control`.

### Закрыть Stage2/P3 одной командой

```bash
# Linux/macOS
./scripts/stage2-p3-closure-check.sh

# либо через Makefile
make stage2-p3-check
```

```powershell
# Windows PowerShell
.\scripts\stage2-p3-closure-check.ps1

# либо через Makefile (если есть make)
make stage2-p3-check-win
```

Сценарий `stage2-p3-closure-check` включает: таргетированные unit-тесты (`internal/cve`, `internal/report`, `internal/remoteexec`, `cmd/network-scanner`), sanity генерации security report в redacted/unredacted режимах и проверку policy-guardrails для `remote-exec`.

### Проверка статуса CI (GitHub Actions)

```bash
# Linux/macOS
./scripts/check-ci-status.sh

# либо через Makefile
make ci-status
```

```powershell
# Windows PowerShell
.\scripts\check-ci-status.ps1

# либо через Makefile (если есть make)
make ci-status-win
```

Скрипт выводит последние прогоны workflow `CI`, показывает URL последнего успешного run (если найден) и проверяет, что required jobs для P3 (`Lint`, `Test*`, `Build and Smoke*`, `Stage2 P1 Closure`, `Stage2 P3 Closure`) полностью зелёные. Для Unix-варианта требуется `python3`.
Скрипт работает в strict-режиме и завершает выполнение с ошибкой, если любой required job не зелёный.

Для быстрой диагностики блокеров перед sign-off (Windows):

```powershell
.\scripts\p0-signoff-preflight.ps1
# или
make p0-preflight-win
```

Единый агрегированный статус Stage2 sign-off (Windows):

```powershell
.\scripts\stage2-signoff-status.ps1
# или
make stage2-signoff-status-win
```

### Запуск CI из консоли (без `gh`)

Перед запуском установите переменную `GITHUB_TOKEN` (token с правами на workflow/repo).

```bash
# Linux/macOS
export GITHUB_TOKEN=ghp_xxx
./scripts/trigger-ci-workflow.sh

# либо через Makefile
make ci-trigger
```

```powershell
# Windows PowerShell
$env:GITHUB_TOKEN = "ghp_xxx"
.\scripts\trigger-ci-workflow.ps1

# либо через Makefile (если есть make)
make ci-trigger-win
```

Скрипт запускает workflow `CI`, ждёт завершения и печатает итоговый `run URL`, `conclusion` и итог валидации required jobs для P3 (`Lint`, `Test*`, `Build and Smoke*`, `Stage2 P1 Closure`, `Stage2 P3 Closure`). Для Unix-варианта требуется `python3`.
Дополнительно выводятся и строго валидируются статусы jobs `Stage2 P1 Closure` и `Stage2 P3 Closure`.

### Автозакрытие P3 sign-off (Windows)

```powershell
# Заполнить P3 Final Sign-off из последнего успешного CI run
.\scripts\finalize-p3-signoff.ps1 -ConfirmedBy "RekadzeAV"

# либо через Makefile (если есть make)
make p3-signoff-win
```

Скрипт обновляет `docs/P3_CLOSURE_CHECKLIST.md` только если найден успешный run workflow `CI` и required jobs (`Lint`, `Test*`, `Build and Smoke*`, `Stage2 P1 Closure`, `Stage2 P3 Closure`) полностью зелёные.

### Автозакрытие P3 sign-off (Linux/macOS)

```bash
# Заполнить P3 Final Sign-off из последнего успешного CI run
./scripts/finalize-p3-signoff.sh RekadzeAV network-scanner ci.yml docs/P3_CLOSURE_CHECKLIST.md RekadzeAV

# либо через Makefile
make p3-signoff
```

Unix-скрипт выполняет те же проверки (`Lint`, `Test*`, `Build and Smoke*`) и обновляет `docs/P3_CLOSURE_CHECKLIST.md` только при полностью зеленом run.

### One-command flow для закрытия P3

```powershell
# Windows PowerShell
$env:GITHUB_TOKEN = "ghp_xxx"
.\scripts\trigger-ci-workflow.ps1
.\scripts\check-ci-status.ps1
.\scripts\finalize-p3-signoff.ps1 -ConfirmedBy "RekadzeAV"
```

```bash
# Linux/macOS
export GITHUB_TOKEN=ghp_xxx
./scripts/trigger-ci-workflow.sh
./scripts/check-ci-status.sh
./scripts/finalize-p3-signoff.sh RekadzeAV network-scanner ci.yml docs/P3_CLOSURE_CHECKLIST.md RekadzeAV
```

Если CI ещё не зелёный, `finalize-p3-signoff` завершится с ошибкой и не изменит чеклист — это ожидаемая защита от преждевременного закрытия.

Для полного цикла одной командой:

```powershell
# Windows
$env:GITHUB_TOKEN = "ghp_xxx"
make p3-close-all-win
```

```bash
# Linux/macOS
export GITHUB_TOKEN=ghp_xxx
make p3-close-all
```

`p3-close-all` делает preflight-проверки: наличие `GITHUB_TOKEN` (обязательно) и `python3` для Unix-ветки.

### Smoke-проверка GUI

Для ручной проверки GUI-режимов используйте чеклист:
- [docs/GUI_SMOKE_CHECKLIST.md](docs/GUI_SMOKE_CHECKLIST.md)

Для быстрой проверки адаптивности GUI на типовых разрешениях и DPI:

```bash
# Linux/macOS
./scripts/smoke-gui-resolution.sh ./network-scanner-gui
```

```powershell
# Windows PowerShell
.\scripts\smoke-gui-resolution.ps1 -GuiExe .\network-scanner-gui.exe
```

Скрипты запускают GUI и печатают матрицу ручной проверки (`1366x768` ... `4K`) и критерии приемки для оконного/полноэкранного режима.

## 📦 Требования

- Go 1.24 или выше
- Для GUI версии требуется C компилятор (GCC) из-за CGO
- Для кросскомпиляции в Windows на macOS/Linux требуется mingw-w64
- Для получения MAC адресов может потребоваться запуск с правами администратора

### Настройка для кросскомпиляции в Windows

Если вы хотите собирать Windows версию на macOS:

1. Установите mingw-w64: `brew install mingw-w64`
2. Проверьте окружение: `./scripts/setup-windows-env.sh`
3. Соберите: `./scripts/build-windows.sh`

Подробнее: [QUICKSTART_WINDOWS_BUILD.md](QUICKSTART_WINDOWS_BUILD.md) или [docs/SETUP_WINDOWS_CROSS_COMPILE.md](docs/SETUP_WINDOWS_CROSS_COMPILE.md)

## 🎯 Основные возможности

- 🔍 Автоматическое определение локальной сети
- 📡 Сканирование активных хостов
- 🔌 Сканирование портов TCP
- 🛠️ Сетевые инструменты: `Ping`, `Traceroute`, `DNS lookup`, `Whois`, `Wi-Fi` (GUI + CLI)
- 🌙 Wake-on-LAN (`WOL`) из CLI и GUI (`Инструменты`)
- 🛡️ Базовый аудит открытых портов (CLI + GUI)
- 🧾 Локальные `Risk Signatures` для "домашних" рисков (CLI + GUI)
- 🎛️ Device Control MVP (HTTP API) с audit trail и confirm для reboot
- 🏷️ Сбор баннеров/версий сервисов для типовых TCP-портов (опционально)
- 🧭 Опциональный SNMP-опрос и построение топологии (`--topology`)
- 🗺️ Экспорт топологии в `json`, `graphml`, `png`, `svg` (для изображений нужен Graphviz `dot`)
- 🖥️ Определение типов устройств
- 📊 Аналитика по протоколам и портам
- 🏷️ Определение производителя по MAC адресу
- 🖼️ GUI-режимы результатов: `Таблица` / `Карточки` с сохранением выбора
- 🧭 GUI-подрежимы результатов: `Devices` / `Security`
- 📋 Табличный вывод с горизонтальной прокруткой и чипами портов
- 🧩 Карточки устройств с адаптивной сеткой (на узком экране 1 колонка)
- 🥧 Аналитика в GUI: markdown-сводка (табличный режим) или 2 круговые диаграммы (карточный режим)
- 🔐 Remote Exec (P3 MVP): `ssh|wmi|winrm` с consent, allowlist policy и audit log

## 🖥️ GUI: режимы отображения результатов

Во вкладке `Сканирование` доступны подрежимы `Devices`/`Security`.
В `Devices` доступны два режима представления:

- `Таблица`:
  - колонки `HostName`, `IP`, `MAC`, `Порты`;
  - порты отображаются как чипы с переносом на новую строку;
  - при нехватке ширины доступен горизонтальный скролл таблицы;
  - под таблицей выводится аналитика по протоколам и типам устройств.
- `Карточки`:
  - сетка карточек устройств (`HostName`, `IP`, `MAC`, порты-чипы);
  - на узком экране сетка переходит в одну колонку;
  - аналитика отображается двумя круговыми диаграммами.

Выбранный режим сохраняется в настройках приложения и автоматически восстанавливается при следующем запуске.
Дополнительно сохраняются сортировка, лимит чипов портов и быстрые фильтры результатов.

### Автопрофиль в GUI

Во вкладке `Сканирование` доступен режим `Автопрофиль сканирования (рекомендуется)`:

- автоматически смягчает тяжелые комбинации `ports`/`threads` для крупных подсетей;
- показывает индикатор состояния (`Автопрофиль: ВКЛ/ВЫКЛ`) в панели параметров и в верхней зоне результатов;
- позволяет открыть пояснение кнопкой `Почему изменены параметры?`.
- кнопка `Рекомендуемые настройки` применяет безопасный профиль в один клик и показывает бейдж `Профиль: ...` с классом (`small/medium/large/very-large`);
- класс рекомендованного профиля сохраняется отдельно в `Preferences` и восстанавливается при следующем запуске GUI (с fallback на legacy текст).

При необходимости полного ручного контроля параметров отключите автопрофиль.

## 🧪 Новые CLI команды для топологии

```bash
# Базовое сканирование + построение топологии (вывод связей в консоль)
./network-scanner --topology

# Построение топологии и сохранение в JSON
./network-scanner --topology --output-format json --output-file topology.json

# Построение топологии и сохранение в GraphML
./network-scanner --topology --output-format graphml --output-file topology.graphml

# Построение топологии и экспорт в PNG (требуется Graphviz/dot)
./network-scanner --topology --output-format png --output-file topology.png

# Несколько SNMP community и увеличенный таймаут
./network-scanner --topology --snmp-community public,private,monitor --snmp-timeout 4
```

Ключевые флаги:
- `--topology` включает построение топологии после обычного сканирования.
- `--output-format` поддерживает `json`, `graphml`, `png`, `svg`.
- `--output-file` задает путь и имя файла для сохранения.
- `--snmp-community` принимает одну или несколько community-строк через запятую.
- `--snmp-timeout` задает SNMP-таймаут в секундах.

## 🛠️ CLI инструменты P1

```bash
# Ping со структурной сводкой
./network-scanner --ping 8.8.8.8

# Ping с кастомным количеством пакетов и таймаутом
./network-scanner --ping 8.8.8.8 --ping-count 6 --tool-timeout 20

# Traceroute со структурной сводкой hop-ов
./network-scanner --traceroute google.com

# Traceroute с ограничением числа hop
./network-scanner --traceroute google.com --traceroute-max-hops 20

# DNS lookup (A/AAAA или PTR)
./network-scanner --dns 1.1.1.1

# DNS lookup через указанный резолвер
./network-scanner --dns example.com --dns-server 1.1.1.1

# Whois lookup (домен или IP)
./network-scanner --whois example.com

# Wi-Fi информация текущей ОС
./network-scanner --wifi

# Показать и структурную сводку, и raw output
./network-scanner --ping 8.8.8.8 --raw
```

Ключевые флаги инструментов:
- `--ping` запускает ping и завершает программу (без полного сканирования сети).
- `--traceroute` запускает traceroute/tracert и завершает программу.
- `--dns` запускает DNS lookup и завершает программу.
- `--whois` запускает whois lookup и завершает программу.
- `--wifi` показывает Wi-Fi информацию текущей ОС и завершает программу.
- `--dns-server` задает DNS сервер для `--dns` (`IP` или `IP:port`).
- `--ping-count` задает число ping-пакетов для `--ping` (1..50).
- `--tool-timeout` задает таймаут (в секундах) для tool-режимов (`--ping`, `--traceroute`, `--dns`, `--whois`, `--wifi`).
- `--traceroute-max-hops` задает максимальное число hop для `--traceroute` (1..64).
- `--raw` дополнительно печатает полный сырой вывод утилиты.

## 🛡️ Аудит открытых портов

```bash
# Базовый аудит рисков после сканирования
./network-scanner --network 192.168.1.0/24 --audit-open-ports

# Показывать только high+ риски
./network-scanner --network 192.168.1.0/24 --audit-open-ports --audit-min-severity high
```

- `--audit-open-ports` анализирует найденные открытые порты по базовым правилам рисков
  (например, Telnet/FTP/SMB/RDP/MongoDB/Redis/Elasticsearch/Memcached).
- `--audit-min-severity` задает минимальную критичность для вывода:
  `all|low|medium|high|critical` (по умолчанию `low`).
- В выводе указываются `severity`, хост/порт и краткая рекомендация по снижению риска.

## 🚀 CLI инструменты P2 (WOL + баннеры/версии)

```bash
# Wake-on-LAN: отправить magic packet на MAC
./network-scanner --wol-mac aa:bb:cc:dd:ee:ff

# Wake-on-LAN: указать конкретный broadcast
./network-scanner --wol-mac aa:bb:cc:dd:ee:ff --wol-broadcast 192.168.1.255:9

# Wake-on-LAN: использовать интерфейс для автоподбора broadcast
./network-scanner --wol-mac aa:bb:cc:dd:ee:ff --wol-iface eth0

# Сканирование с баннерами/версиями
./network-scanner --grab-banners

# Показ raw banner в CLI (по умолчанию скрыт)
./network-scanner --grab-banners --show-raw-banners

# Включить расширенные (active) эвристики определения ОС
./network-scanner --os-detect-active

# Включить подробные логи по каждому порту (для диагностики)
./network-scanner --verbose-port-logs
```

Ключевые флаги P2:
- `--wol-mac` отправляет Wake-on-LAN magic packet и завершает программу.
- `--wol-broadcast` задает broadcast адрес для `--wol-mac` (`IP` или `IP:port`).
- `--wol-iface` задает интерфейс для `--wol-mac`, если `--wol-broadcast` не указан.
- `--grab-banners` включает чтение баннеров/версий сервисов с типовых TCP-портов (может замедлить скан).
- `--show-raw-banners` показывает в CLI сырой banner; без флага отображается только нормализованная версия.
- `--os-detect-active` включает расширенные (active) сигнатуры определения ОС.
- `--verbose-port-logs` включает очень подробные debug-логи по каждому probe порта (рекомендуется только для диагностики).

## 🛡️ CLI инструменты Stage2 P2 (Risk Signatures + Device Control)

```bash
# Risk signatures после сканирования
./network-scanner --network 192.168.1.0/24 --grab-banners --risk-signatures

# Device control: status
./network-scanner \
  --device-action status \
  --device-target http://192.168.1.1 \
  --device-vendor generic-http \
  --device-user admin \
  --device-pass secret

# Device control: reboot (только с явным подтверждением)
./network-scanner \
  --device-action reboot \
  --device-target http://192.168.1.1 \
  --device-vendor tp-link-http \
  --device-user admin \
  --device-pass secret \
  --device-confirm I_UNDERSTAND \
  --audit-log audit/device-actions.log
```

Ключевые флаги Stage2 P2:
- `--risk-signatures` запускает локальные сигнатуры рисков по итогам сканирования.
- `--device-action` запускает tool-режим управления оборудованием (`status`/`reboot`) и завершает программу.
- `--device-target` задает URL API устройства.
- `--device-vendor` выбирает профиль API (`generic-http`, `tp-link-http`).
- `--device-confirm I_UNDERSTAND` обязателен для `--device-action reboot`.
- `--audit-log` задает путь JSONL audit trail для действий device-control.
- `--security-report-file` включает в HTML отчет и CVE, и findings `Risk Signatures`.
- `--security-report-file auto` выбирает имя по режиму и `report-id`: `security-report-redacted-<id>.html` или `security-report-unredacted-<id>.html`.
- `--security-report-redact` управляет маскированием чувствительных данных в HTML security report (`true` по умолчанию).
- Для `--security-report-redact=false` требуется явное подтверждение: `--security-report-unsafe-consent I_UNDERSTAND_UNREDACTED_REPORT`.
- В шапке security report выводится индикатор режима: `REDACTION: ON|OFF`.
- Для unredacted-режима дополнительно выводится warning в CLI и в самом HTML-файле.
- В шапке отчета также выводится metadata-блок: `mode`, `policy`, `unsafe-consent`.

## ⚙️ Ограничения инструментов P2

- WOL работает в пределах L2-сегмента или при специально настроенном directed broadcast/relay.
- Для части устройств нужен ручной enable WOL в BIOS/UEFI и настройках сетевого адаптера.
- Сбор баннеров зависит от поведения сервиса: часть портов может возвращать `нет ответа`.
- HTTPS-баннеры читаются в режиме best-effort (без строгой валидации сертификата).
- Active-режим определения ОС использует дополнительные эвристики по портовым комбинациям и может снижать точность в нестандартных средах.

## ⚙️ Ограничения инструментов P1

- `ping` и `traceroute` используют системные утилиты (`ping`, `tracert`/`traceroute`) и зависят от их наличия в `PATH`.
- Формат raw-вывода зависит от ОС, локали и версии утилиты.
- DNS lookup работает через системный резолвер Go; при `--dns-server` используется указанный сервер.
- `whois` зависит от внешней утилиты `whois` в `PATH`; на Windows может требоваться отдельная установка клиента.
- `--wifi` использует OS-утилиты: `netsh` (Windows), `nmcli` (Linux), `airport` (macOS).
- Сканер хостов по-прежнему определяет «живость» через TCP probe (не ICMP ping).

## 🔐 Remote Exec (P3 MVP)

```bash
# 1) dry-run с policy файлом (рекомендуется)
./network-scanner \
  --remote-exec-transport ssh \
  --remote-exec-target 192.168.1.10 \
  --remote-exec-user admin \
  --remote-exec-command "hostname" \
  --remote-exec-policy-file config/remote-exec-policy.example.json \
  --remote-exec-consent I_UNDERSTAND \
  --remote-exec-dry-run

# 2) реальное выполнение (только после dry-run)
./network-scanner \
  --remote-exec-transport ssh \
  --remote-exec-target 192.168.1.10 \
  --remote-exec-user admin \
  --remote-exec-command "hostname" \
  --remote-exec-policy-file config/remote-exec-policy.example.json \
  --remote-exec-consent I_UNDERSTAND \
  --remote-exec-dry-run=false
```

- Обязательны: `--remote-exec-consent I_UNDERSTAND` + allowlist (`--remote-exec-policy-file` или `--remote-exec-allow-*`).
- По умолчанию включен `--remote-exec-dry-run=true`.
- Для строгого режима используйте `--remote-exec-policy-strict` (только policy-файл, inline allowlist запрещен).
- Для `wmi`/`winrm` запуск поддерживается только на Windows.
- Все операции пишутся в `--remote-exec-audit-log` (JSONL).
- CLI/audit маскируют типовые секреты (`password/token/secret/api-key`) в сообщениях и выводе.

## 📝 Лицензия

Этот проект распространяется под лицензией MIT. См. файл [LICENSE](LICENSE) для подробностей.

## 🤝 Вклад в проект

Проект открыт для вклада! Если вы хотите улучшить проект:

1. Создайте форк репозитория
2. Создайте ветку для вашей функции (`git checkout -b feature/AmazingFeature`)
3. Зафиксируйте изменения (`git commit -m 'Add some AmazingFeature'`)
4. Отправьте в ветку (`git push origin feature/AmazingFeature`)
5. Откройте Pull Request

Подробные правила: [CONTRIBUTING.md](CONTRIBUTING.md)

## 📋 История изменений

См. [CHANGELOG.md](CHANGELOG.md) для списка всех изменений в проекте.

## ⚠️ Предупреждение

Этот инструмент предназначен для использования только в ваших собственных сетях или сетях, где у вас есть явное разрешение на сканирование. Не используйте его для несанкционированного сканирования сетей.

## 🔗 Ссылки

- [Инструкция по эксплуатации](Инструкция%20по%20эксплуатации.md) - Полная инструкция по эксплуатации (русский язык)
- [Руководство пользователя](docs/USER_GUIDE.md) - Подробное руководство пользователя
- [Инструкция по установке](docs/INSTALL.md) - Инструкции по установке
- [Техническая документация](docs/TECHNICAL.md) - Техническая документация
- [Архитектура проекта](docs/ARCHITECTURE.md) - Архитектура проекта
- [Анализ проекта](docs/ANALYSIS.md) - Анализ проекта
- [Карта разработки](DEVELOPMENT_MAP.md) - Детальная карта разработки
- [История изменений](CHANGELOG.md) - История изменений
- [Быстрый старт: Windows сборка](QUICKSTART_WINDOWS_BUILD.md) - Сборка для Windows на macOS

---

**Версия документа:** 1.0.5  
**Последнее обновление:** 2026-04-23
