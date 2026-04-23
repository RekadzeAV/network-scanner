# Release Readiness Snapshot

Краткий срез готовности релиза на текущий момент.

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
- Security report sanity:
  - генерация `release/security-report-sanity.html`;
  - подтверждены секции `CVE Findings` и `Risk Signature Findings`.

## Документация синхронизирована

- `README.md`
- `docs/USER_GUIDE.md`
- `docs/TECHNICAL.md`
- `docs/GUI_SMOKE_CHECKLIST.md`
- `docs/RELEASE_SUMMARY_UI_RESULTS.md`
- `docs/PR_DESCRIPTION_UI_RESULTS.md`
- `CHANGELOG.md`
- `docs/RELEASE_ACCEPTANCE_CHECKLIST.md`
- `docs/P1_CLOSURE_CHECKLIST.md`

## Что осталось до финального sign-off

### Ручная GUI-приемка

- Пройти `docs/GUI_SMOKE_CHECKLIST.md`.
- Подтвердить:
  - сохранение/восстановление режима, сортировки, фильтров и лимита чипов;
  - корректный экспорт текущего представления (`Сохранить результаты`);
  - корректную работу Stage2 P2 инструментов (`Risk Signatures`, `Device Control`, audit-лог).

### Ручные проверки топологии

- Вкладка `Топология`: открыть/построить/сохранить/масштабирование превью.
- Внешняя совместимость `GraphML`:
  - import в yEd;
  - import в Gephi;
  - фиксация по `docs/GRAPHML_COMPATIBILITY_CHECK.md`.

### CI evidence

- Получить успешный GitHub Actions run workflow `CI` (`Lint`, `Test`, `Build and Smoke`).
- Подтвердить прохождение `Windows/Linux/macOS`.
- Внести URL успешного run в `docs/P3_CLOSURE_CHECKLIST.md` (раздел `P3 Final Sign-off`).

## Быстрый чек перед выпуском

1. Закрыть ручные пункты в `docs/RELEASE_ACCEPTANCE_CHECKLIST.md`.
2. Зафиксировать CI evidence.
3. Убедиться, что версия/дата в релевантных docs актуальны.
4. Подготовить финальные release notes и PR описание (черновики уже готовы).

