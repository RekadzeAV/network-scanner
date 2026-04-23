# Manual Sign-off Draft

Черновик заполнен на основании уже выполненных автоматических проверок в текущей ветке.
Осталось внести результаты ручной GUI/Topology/GraphML валидации.

## Context

- Reviewer: `<to fill>`
- Date: `2026-04-23`
- Environment (OS/arch): `Windows amd64`
- Build under test (commit/tag): `<to fill>`

## GUI Smoke (`docs/GUI_SMOKE_CHECKLIST.md`)

- [ ] Пройден полностью
- Notes:
  - Сканирование/остановка/состояния: `<to fill>`
  - Режимы `Таблица`/`Карточки`: `<to fill>`
  - Автопрофиль (`ВКЛ/ВЫКЛ`, индикаторы, инфо-диалог): `<to fill>`
  - Персистентность настроек: `<to fill>`
  - Экспорт текущего представления: `<to fill>`
  - Инструменты (`Ping`/`Traceroute` raw output, Windows ping summary): `<to fill>`

## Topology Manual Checks

- [ ] Вкладка `Топология` открывается корректно
- [ ] Построение топологии без ошибок
- [ ] Превью PNG/масштабирование/сохранение работают
- Notes:
  - SNMP summary: `<to fill>`
  - Build duration / responsiveness: `<to fill>`
  - Preview/save behavior: `<to fill>`

## GraphML Compatibility (`docs/GRAPHML_COMPATIBILITY_CHECK.md`)

- [ ] yEd import verified
- [ ] Gephi import verified
- [ ] Results captured in compatibility checklist/doc
- Notes:
  - yEd: `<to fill>`
  - Gephi: `<to fill>`

## CI Evidence

- [ ] `CI` workflow green (`Lint`, `Test`, `Build and Smoke`)
- [ ] Windows/Linux/macOS confirmed
- [ ] CI URL recorded in `docs/P3_CLOSURE_CHECKLIST.md`
- CI run URL: `<to fill>`

## Already Confirmed (Auto Evidence)

- [x] `go test ./...` pass
- [x] `go test -tags=integration ./...` pass
- [x] CLI build: `go build -o network-scanner ./cmd/network-scanner`
- [x] GUI build: `go build -o network-scanner-gui ./cmd/gui`
- [x] CLI smoke:
  - [x] `scripts/smoke-cli-no-topology.ps1`
  - [x] `scripts/smoke-cli-topology.ps1`
  - [x] `scripts/smoke-cli-tools.ps1`
  - [x] `scripts/smoke-d-track-topology-export.ps1`
- [x] Closure scripts:
  - [x] `scripts/p1-closure-check.ps1`
  - [x] `scripts/p2-closure-check.ps1`
  - [x] `scripts/p3-closure-check.ps1`
- [x] Security report sanity:
  - [x] `release/security-report-sanity.html` generated
  - [x] sections present: `CVE Findings`, `Risk Signature Findings`

## Final Decision

- [ ] Ready for final sign-off
- [ ] Defer (follow-up required)
- Residual risks / follow-ups: `<to fill>`

