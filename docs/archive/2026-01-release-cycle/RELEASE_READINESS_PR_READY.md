# Release Readiness PR Block (Ready-to-Paste)

Ниже даны short/long варианты на английском и русском.

## Short Version

Release readiness is largely confirmed on local Windows: `go test ./...` and `go test -tags=integration ./...` pass, CLI/GUI builds succeed, smoke + closure scripts are green, and security report sanity is validated (`CVE Findings` + `Risk Signature Findings` present).  
Remaining items before final sign-off: manual GUI/topology checks, external GraphML compatibility (yEd/Gephi), and CI evidence across Windows/Linux/macOS with run URL recorded in [P3_CLOSURE_CHECKLIST.md](P3_CLOSURE_CHECKLIST.md).

## Короткая версия (RU)

Готовность релиза в основном подтверждена локально на Windows: `go test ./...` и `go test -tags=integration ./...` проходят, CLI/GUI собираются, smoke + closure-скрипты зелёные, sanity security report подтверждён (в отчёте есть секции `CVE Findings` и `Risk Signature Findings`).  
До финального sign-off осталось закрыть ручные проверки GUI/топологии, внешнюю проверку совместимости GraphML (yEd/Gephi) и добавить CI evidence по Windows/Linux/macOS с URL успешного прогона в [P3_CLOSURE_CHECKLIST.md](P3_CLOSURE_CHECKLIST.md).

## Long Version

### Current Status

Release readiness is largely confirmed on local Windows environment: unit/integration tests pass, CLI/GUI builds succeed, smoke/closure scripts are green, and security report sanity checks are validated.

### Completed (Evidence)

- Tests:
  - `go test ./...` passed
  - `go test -tags=integration ./...` passed
- Builds:
  - `go build -o network-scanner ./cmd/network-scanner`
  - `go build -o network-scanner-gui ./cmd/gui`
- Smoke scripts:
  - `scripts/smoke-cli-no-topology.ps1`
  - `scripts/smoke-cli-topology.ps1`
  - `scripts/smoke-cli-tools.ps1`
  - `scripts/smoke-d-track-topology-export.ps1`
- Closure scripts:
  - `scripts/p1-closure-check.ps1`
  - `scripts/p2-closure-check.ps1`
  - `scripts/p3-closure-check.ps1`
- Security report sanity:
  - generated `build/release/security-report-sanity.html`
  - verified sections: `CVE Findings`, `Risk Signature Findings`

### Docs Synced

- [README.md](../README.md)
- [USER_GUIDE.md](USER_GUIDE.md)
- [TECHNICAL.md](TECHNICAL.md)
- [GUI_SMOKE_CHECKLIST.md](GUI_SMOKE_CHECKLIST.md)
- [RELEASE_SUMMARY_UI_RESULTS.md](RELEASE_SUMMARY_UI_RESULTS.md)
- [PR_DESCRIPTION_UI_RESULTS.md](PR_DESCRIPTION_UI_RESULTS.md)
- [RELEASE_ACCEPTANCE_CHECKLIST.md](RELEASE_ACCEPTANCE_CHECKLIST.md)
- [P1_CLOSURE_CHECKLIST.md](P1_CLOSURE_CHECKLIST.md)
- [CHANGELOG.md](../CHANGELOG.md)

### Remaining Before Final Sign-off

- Manual GUI acceptance ([GUI_SMOKE_CHECKLIST.md](GUI_SMOKE_CHECKLIST.md))
- Manual topology/preview checks in GUI
- External GraphML compatibility checks (yEd/Gephi)
- CI evidence:
  - successful GitHub Actions `CI` run (`Lint`, `Test`, `Build and Smoke`)
  - confirmation on Windows/Linux/macOS
  - run URL added to [P3_CLOSURE_CHECKLIST.md](P3_CLOSURE_CHECKLIST.md) (`P3 Final Sign-off`)

## Расширенная версия (RU)

### Текущий статус

Готовность релиза в основном подтверждена в локальной Windows-среде: unit/integration тесты проходят, CLI/GUI сборки успешны, smoke/closure-скрипты зелёные, sanity security report валиден.

### Что подтверждено (Evidence)

- Тесты:
  - `go test ./...` — успешно
  - `go test -tags=integration ./...` — успешно
- Сборка:
  - `go build -o network-scanner ./cmd/network-scanner`
  - `go build -o network-scanner-gui ./cmd/gui`
- Smoke-скрипты:
  - `scripts/smoke-cli-no-topology.ps1`
  - `scripts/smoke-cli-topology.ps1`
  - `scripts/smoke-cli-tools.ps1`
  - `scripts/smoke-d-track-topology-export.ps1`
- Closure-скрипты:
  - `scripts/p1-closure-check.ps1`
  - `scripts/p2-closure-check.ps1`
  - `scripts/p3-closure-check.ps1`
- Security report sanity:
  - сгенерирован `build/release/security-report-sanity.html`
  - подтверждены секции: `CVE Findings`, `Risk Signature Findings`

### Синхронизированная документация

- [README.md](../README.md)
- [USER_GUIDE.md](USER_GUIDE.md)
- [TECHNICAL.md](TECHNICAL.md)
- [GUI_SMOKE_CHECKLIST.md](GUI_SMOKE_CHECKLIST.md)
- [RELEASE_SUMMARY_UI_RESULTS.md](RELEASE_SUMMARY_UI_RESULTS.md)
- [PR_DESCRIPTION_UI_RESULTS.md](PR_DESCRIPTION_UI_RESULTS.md)
- [RELEASE_ACCEPTANCE_CHECKLIST.md](RELEASE_ACCEPTANCE_CHECKLIST.md)
- [P1_CLOSURE_CHECKLIST.md](P1_CLOSURE_CHECKLIST.md)
- [CHANGELOG.md](../CHANGELOG.md)

### Что осталось до финального sign-off

- Ручная GUI-приёмка ([GUI_SMOKE_CHECKLIST.md](GUI_SMOKE_CHECKLIST.md))
- Ручные проверки топологии/превью в GUI
- Внешние проверки совместимости GraphML (yEd/Gephi)
- CI evidence:
  - успешный GitHub Actions workflow `CI` (`Lint`, `Test`, `Build and Smoke`)
  - подтверждение прохождения на Windows/Linux/macOS
  - добавление URL успешного прогона в [P3_CLOSURE_CHECKLIST.md](P3_CLOSURE_CHECKLIST.md) (`P3 Final Sign-off`)

## PR Final Comment Checklist (EN)

- [ ] Acceptance checklist updated ([RELEASE_ACCEPTANCE_CHECKLIST.md](RELEASE_ACCEPTANCE_CHECKLIST.md))
- [ ] CI run is green for `Lint`, `Test`, `Build and Smoke`
- [ ] Manual GUI smoke/topology checks completed
- [ ] GraphML compatibility verified (yEd/Gephi) or explicitly deferred
- [ ] Final PR comment includes evidence links/paths and remaining risks

## Чеклист финального комментария в PR (RU)

- [ ] Обновлен acceptance checklist ([RELEASE_ACCEPTANCE_CHECKLIST.md](RELEASE_ACCEPTANCE_CHECKLIST.md))
- [ ] CI зелёный по `Lint`, `Test`, `Build and Smoke`
- [ ] Выполнены ручные GUI smoke/topology проверки
- [ ] Проверена совместимость GraphML (yEd/Gephi) или явно отмечен defer
- [ ] Финальный комментарий в PR содержит evidence-ссылки/пути и остаточные риски

