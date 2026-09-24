# Назначение рабочих веток (2026-09-25)

Маркер того, какой поток работ ведётся в какой ветке. Дублирует процесс из
[../../CONTRIBUTING.md](../../CONTRIBUTING.md) (trunk-based, Conventional Commits)
и roadmap из [../ROADMAP.md](../ROADMAP.md).

## `feature/v2.3.0-plan-continuation` — реализация плана v2.3.0

**Назначение:** продолжение работ по каноническому плану
[../UNIFIED_OPTIMIZED_PLAN_2026-09-15.md](../UNIFIED_OPTIMIZED_PLAN_2026-09-15.md)
(этапы E0–E7) и по открытым пунктам [../ROADMAP.md](../ROADMAP.md).

**Что уже реализовано в этой ветке** (база — коммит `b6a7c90`):

| Направление | Содержание |
|-------------|------------|
| M2 — гонки данных | `internal/ports`: shared `cases.Caser` заменён чистой `titleCaseWord` (устранён DATA RACE на пути `scanHost → GetServiceName → LookupServiceName`); `internal/network`: `TestARPCacheRapidRefresh` переведён на `atomic` |
| Регрессионные тесты | `TestLookupServiceName_Concurrent`, `TestLookupServiceName_ConcurrentParallel`, `TestTitleCaseWord`, `TestFormatIANAServiceName_Whitespace` |
| Инструменты | Makefile-цель `test-race` (`go test -race ./... -short`) |
| Документация | CHANGELOG (M2), ROADMAP (M2 + race-статус), UNIFIED-план (E4/E5/E6 закрыты), README (метрики), исправлена ссылка в `docs/dev/README.md` |
| Проверки | `go build ./...`, `go vet ./...`, `golangci-lint run ./...` (0 замечаний), `go test ./... -short` (52 пакета ok), `go test -race ./... -short` (0 FAIL, 0 DATA RACE), `docs-link-check` |

**Следующие шаги в этой ветке:** CI-job `race`, затем расширения E7 (ICMP ping
тесты/документация, UDP-скан, PDF-отчёты, plugin-probes, event-driven GUI) и P3
(метрики Prometheus, strict TLS в remoteexec, аудит-лог изменяющих операций).

## `feature/gui-design` — переработка GUI

**Назначение:** UI/UX-работа по Fyne v2.7.1: декомпозиция и компоновка панелей,
фильтры-accordion, пресеты, статус-бар/тосты, валидация ввода, результаты/аналитика.

**История в ветке:** Этап 1–3 UX-плана (`b0b7074`, `ec786d5`, `86c36bd`),
декомпозиция `buildResultsContainer` на секции (R10, `9864d55`).

**Правило:** изменения только GUI-слоя; изменения ядра/CLI/API — через
`feature/v2.3.0-plan-continuation`, чтобы не смешивать потоки работ.

## `main`

**Назначение:** интеграционная ветка (trunk). Принимает завершённые изменения из
рабочих веток короткими PR после зелёного CI (lint, test, build, govulncheck).
Сейчас опережает `origin/main` на 1 коммит (`b6a7c90`).

## Правила

1. Одна ветка = один поток работ (реализация плана ≠ UI-переработка).
2. Перед PR — локальные гейты: `go build ./...`, `go vet ./...`,
   `golangci-lint run ./...`, `go test ./... -short`, `go test -race ./... -short`.
3. `main` не коммитится напрямую для крупных изменений — только merge/PR из рабочих веток.
