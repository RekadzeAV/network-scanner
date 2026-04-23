# Release Readiness Snapshot

Краткий срез готовности релиза на текущий момент.

## Готовность фаз Этапа 1

- `P1` — **100%** (локальный closure `scripts/p1-closure-check.ps1` пройден).
- `P2` — **100%** (локальный closure `scripts/p2-closure-check.ps1` пройден).
- `P3` — **100%** (локальный closure `scripts/p3-closure-check.ps1` пройден).

## Готовность фаз Этапа 2

- `P1` — **100%** (локальный closure `scripts/stage2-p1-closure-check.ps1` пройден).
- `P2` — **100%** (локальный closure `scripts/stage2-p2-closure-check.ps1` пройден).
- `P3` — **100%** (локальный closure `scripts/stage2-p3-closure-check.ps1` пройден).

## Автоматически подтверждено (локально, Windows)

- `go test ./...` — пройден.
- `go test -tags=integration ./...` — пройден.
- Сборка артефактов:
  - `go build -o network-scanner ./cmd/network-scanner`
  - `go build -o network-scanner-gui ./cmd/gui`
- Smoke-скрипты CLI:
  - `scripts/smoke-cli-no-topology.ps1`
  - `scripts/smoke-cli-topology.ps1`
  - `scripts/smoke-cli-tools.ps1`
  - `scripts/smoke-d-track-topology-export.ps1`
- Closure-скрипты:
  - `scripts/p1-closure-check.ps1`
  - `scripts/p2-closure-check.ps1`
  - `scripts/p3-closure-check.ps1`
  - `scripts/stage2-p1-closure-check.ps1`
  - `scripts/stage2-p2-closure-check.ps1`
  - `scripts/stage2-p3-closure-check.ps1`
- Security report sanity:
  - генерация `build/release/security-report-sanity.html`;
  - подтверждены секции `CVE Findings` и `Risk Signature Findings`.
  - подтверждены guardrails unredacted-режима и auto-именование report-файлов.

## Документация синхронизирована

- [README.md](../README.md)
- [USER_GUIDE.md](USER_GUIDE.md)
- [TECHNICAL.md](TECHNICAL.md)
- [GUI_SMOKE_CHECKLIST.md](GUI_SMOKE_CHECKLIST.md)
- [RELEASE_SUMMARY_UI_RESULTS.md](RELEASE_SUMMARY_UI_RESULTS.md)
- [PR_DESCRIPTION_UI_RESULTS.md](PR_DESCRIPTION_UI_RESULTS.md)
- [CHANGELOG.md](../CHANGELOG.md)
- [RELEASE_ACCEPTANCE_CHECKLIST.md](RELEASE_ACCEPTANCE_CHECKLIST.md)
- [P1_CLOSURE_CHECKLIST.md](P1_CLOSURE_CHECKLIST.md)

## Что осталось до финального sign-off

### Ручная GUI-приемка

- Пройти [GUI_SMOKE_CHECKLIST.md](GUI_SMOKE_CHECKLIST.md).
- Подтвердить:
  - сохранение/восстановление режима, сортировки, фильтров и лимита чипов;
  - корректный экспорт текущего представления (`Сохранить результаты`);
  - корректную работу Stage2 P2 инструментов (`Risk Signatures`, `Device Control`, audit-лог).

### Ручные проверки топологии

- Вкладка `Топология`: открыть/построить/сохранить/масштабирование превью.
- Внешняя совместимость `GraphML`:
  - import в yEd;
  - import в Gephi;
  - фиксация по [GRAPHML_COMPATIBILITY_CHECK.md](GRAPHML_COMPATIBILITY_CHECK.md).

### CI evidence

- Получить успешный GitHub Actions run workflow `CI` (`Lint`, `Test`, `Build and Smoke`).
- Подтвердить прохождение `Windows/Linux/macOS`.
- Внести URL успешного run в [P3_CLOSURE_CHECKLIST.md](P3_CLOSURE_CHECKLIST.md) (раздел `P3 Final Sign-off`).
- Текущий preflight-статус (Windows): `BLOCKED` (`GITHUB_TOKEN` отсутствует, нет успешного recent `ci.yml` run, нет рабочего `bash/sh` runtime для Unix closure в текущей среде).
- Быстрая проверка перед sign-off: `.\scripts\p0-signoff-preflight.ps1` (или `make p0-preflight-win`).

## Быстрый чек перед выпуском

1. Закрыть ручные пункты в [RELEASE_ACCEPTANCE_CHECKLIST.md](RELEASE_ACCEPTANCE_CHECKLIST.md).
2. Зафиксировать CI evidence.
3. Убедиться, что версия/дата в релевантных docs актуальны.
4. Подготовить финальные release notes и PR описание (черновики уже готовы).

## Short gap list

Для оперативной работы по оставшимся задачам использовать: [RELEASE_READY_GAP_LIST.md](RELEASE_READY_GAP_LIST.md).
Сводный индекс по всем checklist-документам: [CHECKLIST_STATUS_INDEX.md](CHECKLIST_STATUS_INDEX.md).

