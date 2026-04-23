# Release Acceptance Checklist

Финальная проверка перед выпуском версии.

## Evidence quick links

- Release snapshot: [RELEASE_READINESS_SNAPSHOT.md](RELEASE_READINESS_SNAPSHOT.md)
- Stage2/P3 summary: [RELEASE_SUMMARY_STAGE2_P3.md](RELEASE_SUMMARY_STAGE2_P3.md)
- PR snippets (EN/RU): [RELEASE_READINESS_PR_SNIPPET.md](RELEASE_READINESS_PR_SNIPPET.md)
- Stage2/P3 PR snippet (EN/RU): [RELEASE_READINESS_STAGE2_P3_PR_SNIPPET.md](RELEASE_READINESS_STAGE2_P3_PR_SNIPPET.md)
- Stage2/P3 final PR comment (RU): [FINAL_PR_COMMENT_STAGE2_P3_READY.md](FINAL_PR_COMMENT_STAGE2_P3_READY.md)
- PR ready blocks (short/long EN+RU): [RELEASE_READINESS_PR_READY.md](RELEASE_READINESS_PR_READY.md)
- Final PR comment (ready): [FINAL_PR_COMMENT_READY.md](FINAL_PR_COMMENT_READY.md)
- Manual sign-off template: [MANUAL_SIGNOFF_TEMPLATE.md](MANUAL_SIGNOFF_TEMPLATE.md)
- Manual sign-off draft (pre-filled): [MANUAL_SIGNOFF_DRAFT.md](MANUAL_SIGNOFF_DRAFT.md)
- Short gap list: [RELEASE_READY_GAP_LIST.md](RELEASE_READY_GAP_LIST.md)
- Checklist status index: [CHECKLIST_STATUS_INDEX.md](CHECKLIST_STATUS_INDEX.md)
- P0 sign-off runbook: [P0_SIGNOFF_RUNBOOK.md](P0_SIGNOFF_RUNBOOK.md)
- Release operations (closure commands + локальные бинарники в `build/release/`): [RELEASE_OPERATIONS_CHEATSHEET.md](RELEASE_OPERATIONS_CHEATSHEET.md)
- Структура каталогов релизной сборки: [BUILD_STRUCTURE.md](BUILD_STRUCTURE.md)

## 1) Репозиторий и документация

- [x] `CHANGELOG.md` обновлен (раздел `Unreleased` или целевая версия).
- [x] `README.md` отражает актуальные возможности и ссылки на новые документы.
- [x] Обновлены пользовательские документы:
  - [x] [USER_GUIDE.md](USER_GUIDE.md)
  - [x] [GUI_SMOKE_CHECKLIST.md](GUI_SMOKE_CHECKLIST.md)
  - [x] [RELEASE_SUMMARY_UI_RESULTS.md](RELEASE_SUMMARY_UI_RESULTS.md)
  - [x] [PR_DESCRIPTION_UI_RESULTS.md](PR_DESCRIPTION_UI_RESULTS.md)

## 2) Автотесты и качество

- [x] Выполнен `go test ./...` без ошибок.
- [x] Выполнен `go test -tags=integration ./...` без ошибок.
- [x] Нет новых проблем линтера в измененных файлах.
- [x] Проверены ключевые unit-тесты модели результатов:
  - [x] сортировка;
  - [x] фильтрация;
  - [x] нормализация типов;
  - [x] лимит/формат чипов портов.
- [x] Пройден локальный `P3` closure-check:
  - Windows: `.\scripts\p3-closure-check.ps1` (или `make p3-check-win`)
  - Linux/macOS: `./scripts/p3-closure-check.sh` (или `make p3-check`)

## 3) Сборка артефактов

- [x] CLI собирается: `go build -o network-scanner ./cmd/network-scanner`
- [x] GUI собирается: `go build -o network-scanner-gui ./cmd/gui`
- [x] При необходимости выполнены smoke-скрипты CLI:
  - `scripts/smoke-cli-no-topology.sh|ps1`
  - `scripts/smoke-cli-topology.sh|ps1`
  - `scripts/smoke-cli-tools.sh|ps1`
  - `scripts/smoke-d-track-topology-export.sh|ps1`

### 3.1) Кросс-ОС прогон (Stage 1 P2/P1)

- [x] Windows: выполнен локальный closure-прогон (`.\scripts\p2-closure-check.ps1` / `make p2-check-win`).
- [ ] Linux: выполнены команды:
  - `go test ./...`
  - `./scripts/smoke-cli-no-topology.sh`
  - `./scripts/smoke-cli-topology.sh`
  - `./scripts/smoke-cli-tools.sh`
  - `./scripts/p1-closure-check.sh`
  - `./scripts/p2-closure-check.sh` (или `make p2-check`)
- [ ] macOS: выполнены команды:
  - `go test ./...`
  - `./scripts/smoke-cli-no-topology.sh`
  - `./scripts/smoke-cli-topology.sh`
  - `./scripts/smoke-cli-tools.sh`
  - `./scripts/p1-closure-check.sh`
  - `./scripts/p2-closure-check.sh` (или `make p2-check`)

Текущий операционный статус:
- Windows: ✅ подтверждено локально.
- Linux: ⏳ ожидает прогон в целевой среде.
- macOS: ⏳ ожидает прогон в целевой среде.

### 3.1.1) Кросс-ОС прогон (Stage 2 P1: Whois/Wi-Fi/Audit)

- [x] Windows: выполнен локальный closure-прогон (`.\scripts\stage2-p1-closure-check.ps1` / `make stage2-p1-check-win`).
- [ ] Linux/macOS: выполнен closure-прогон:
  - `./scripts/stage2-p1-closure-check.sh` (или `make stage2-p1-check`)

### 3.2) CI evidence для P3 sign-off

- [ ] Получен успешный GitHub Actions run workflow `CI` (jobs `Lint`, `Test`, `Build and Smoke`).
- [ ] Подтверждено прохождение на `Windows/Linux/macOS`.
- [ ] URL успешного CI run внесен в [P3_CLOSURE_CHECKLIST.md](P3_CLOSURE_CHECKLIST.md) (`P3 Final Sign-off`).
- [ ] Job `Stage2 P1 Closure` в workflow `CI` прошел успешно (Ubuntu + Windows).
- [ ] Job `Stage2 P3 Closure` в workflow `CI` прошел успешно (Ubuntu + Windows).

Текущее состояние:
- Blocked: отсутствует `GITHUB_TOKEN` в окружении для автоматического `p3-close-all`.
- Последние доступные runs `ci.yml` имеют `failure`, требуется новый green-run.

Quick unblock (Windows PowerShell):

```powershell
.\scripts\p0-signoff-preflight.ps1
$env:GITHUB_TOKEN = "<your_token>"
make p3-close-all-win
```

### 3.3) Stage2 P2 closure-check

- [x] Windows: `.\scripts\stage2-p2-closure-check.ps1` (или `make stage2-p2-check-win`) проходит.
- [ ] Linux/macOS: `./scripts/stage2-p2-closure-check.sh` (или `make stage2-p2-check`) проходит.

### 3.4) Stage2 P3 closure-check (CVE/Report/Remote Exec)

- [x] Windows: `.\scripts\stage2-p3-closure-check.ps1` (или `make stage2-p3-check-win`) проходит.
- [ ] Linux/macOS: `./scripts/stage2-p3-closure-check.sh` (или `make stage2-p3-check`) проходит.

## 4) GUI приёмка (ручная)

- [ ] Пройден [GUI_SMOKE_CHECKLIST.md](GUI_SMOKE_CHECKLIST.md).
- [ ] Проверено сохранение/восстановление:
  - подрежима (`Devices`/`Security`);
  - режима (`Таблица`/`Карточки`);
  - сортировки;
  - фильтров;
  - лимита чипов.
- [ ] Проверено, что `Сохранить результаты` экспортирует текущее представление (фильтры + сортировка).
- [ ] Проверен `Host Details Drawer`:
  - корректный выбор хоста из таблицы/карточек;
  - отображение деталей хоста;
  - быстрые действия (`Ping`/`Traceroute`/`DNS`/`Whois`/`WOL`).
- [ ] Проверен подрежим `Security`:
  - отображается `Security Dashboard` (сводка + таблица findings);
  - работает `Export security report (HTML)`.
- [ ] Проверены инструменты Stage2 P2 на вкладке `Инструменты`:
  - `Risk Signatures` показывает explain/findings по текущим результатам;
  - `Device Control` (status/reboot) работает для тестового endpoint;
  - для `reboot` появляется подтверждение опасного действия;
  - действия пишутся в `audit/device-actions.log`.
- [ ] Проверены инструменты Stage2 P1 на вкладке `Инструменты`:
  - `Whois` работает при установленной утилите `whois` и через RDAP fallback;
  - `Wi-Fi` возвращает summary + raw output;
  - `Аудит портов` учитывает `Audit min severity` (`all|low|medium|high|critical`).
- [ ] Проверен `Operations Center`:
  - история операций обновляется после запусков;
  - `Retry` доступен для `failed/canceled`;
  - `Cancel` доступен для `running`.

## 5) Регрессии топологии

- [ ] Вкладка `Топология` открывается корректно.
- [ ] Построение топологии отрабатывает без ошибок.
- [ ] Превью PNG/масштабирование/сохранение работают.
- [x] Проверен D-track smoke экспорта топологии:
  - [x] эквивалентность `json`/`graphml` (узлы/связи);
  - [x] корректный fallback в `json` при отсутствии Graphviz (`png/svg`).
- [ ] Выполнена ручная проверка совместимости `GraphML` во внешних инструментах:
  - yEd import;
  - Gephi import;
  - подтверждение согласно [GRAPHML_COMPATIBILITY_CHECK.md](GRAPHML_COMPATIBILITY_CHECK.md).

## 6) Релизные артефакты

- [x] Подготовлен текст релиз-нотов (можно использовать [RELEASE_SUMMARY_UI_RESULTS.md](RELEASE_SUMMARY_UI_RESULTS.md)).
- [x] Подготовлено описание PR (можно использовать [PR_DESCRIPTION_UI_RESULTS.md](PR_DESCRIPTION_UI_RESULTS.md)).
- [x] Проверена версия/дата в релевантных документах.
- [x] Проверен security report (`--security-report-file`) на наличие секций:
  - [x] `CVE Findings`;
  - [x] `Risk Signature Findings`.
- [x] Выполнен локальный sanity-run генерации security report:
  - `go run ./cmd/network-scanner --network 127.0.0.1/32 --ports 22,80 --timeout 1 --threads 10 --risk-signatures --security-report-file build/release/security-report-sanity.html`
- [x] Проверен guardrail unredacted-режима:
  - без `--security-report-unsafe-consent I_UNDERSTAND_UNREDACTED_REPORT` запуск с `--security-report-redact=false` завершается ошибкой;
  - в HTML отчете присутствуют `REDACTION: OFF` и warning-блок.
- [x] Проверено auto-именование security report:
  - `--security-report-file auto --security-report-redact=true` -> `security-report-redacted-<report-id>.html`;
  - `--security-report-file auto --security-report-redact=false ...unsafe-consent...` -> `security-report-unredacted-<report-id>.html`;
  - в CLI выводится `report-id`.
