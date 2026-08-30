# Roadmap

This file is the canonical roadmap entry point for the project.

## Текущее состояние проекта

**Версия:** v2.2  
**Базовый функционал:** ✅ 100% (13/13 задач)  
**v2.1 Backlog:** ✅ 28/28 задач выполнено (100%)  
**v2.2 Backlog:** ✅ 13/13 задач выполнено (100%)  
**Фокус развития:** Тесты бизнес-логики, интеграционные тесты, CI/CD  
**Дата обновления:** 2026-08-29  

### Текущее покрытие тестами

| Пакет | Coverage | Статус |
|-------|----------|--------|
| `internal/gui` | 51.6% | ✅ Стабильно (integration tests) |
| `internal/gui/controller` | 44.4% | ✅ Стабильно |
| `internal/gui/errors` | 90.7% | ✅ Стабильно |
| `internal/audit` | 99.1% | ✅ |
| `internal/cve` | 100.0% | ✅ |
| `internal/builder` | 100.0% | ✅ |
| `internal/scanner` | 71.9% | 🚧 В процессе |
| `internal/topology` | 87.0% | ✅ |
| `internal/network` | 85.3% | ✅ |
| `internal/plugin` | 100.0% | ✅ |
| `internal/api` | 85.0% | ✅ |

---

## Текущий план

- Основной план реализации v2.0: [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md)
- Детальные приоритеты и вехи: [ROADMAP_P1_P3.md](ROADMAP_P1_P3.md)
- Детализированный бэклог задач: [DETAILED_BACKLOG_P3_STAGE2.md](DETAILED_BACKLOG_P3_STAGE2.md)
- **v2.1 Backlog:** [TASK_BACKLOG_V21.md](TASK_BACKLOG_V21.md) — 27/28 задач выполнено

## Release и closure операции

- Release acceptance checklist: [RELEASE_ACCEPTANCE_CHECKLIST.md](RELEASE_ACCEPTANCE_CHECKLIST.md)
- Readiness snapshot: [RELEASE_READINESS_SNAPSHOT.md](RELEASE_READINESS_SNAPSHOT.md)
- Prioritized gap list: [RELEASE_READY_GAP_LIST.md](RELEASE_READY_GAP_LIST.md)
- Checklist status index: [CHECKLIST_STATUS_INDEX.md](CHECKLIST_STATUS_INDEX.md)
- P0 sign-off runbook: [P0_SIGNOFF_RUNBOOK.md](P0_SIGNOFF_RUNBOOK.md)
- Stage2 commit-ready summary: [STAGE2_100_COMMIT_READY.md](STAGE2_100_COMMIT_READY.md)
- Runbook commands and local release build output: [RELEASE_OPERATIONS_CHEATSHEET.md](RELEASE_OPERATIONS_CHEATSHEET.md) (artifacts under `build/release/`)

## Maintenance rules

- Update this file when roadmap structure changes.
- Keep [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) updated after each completed phase.
- Keep [ROADMAP_P1_P3.md](ROADMAP_P1_P3.md) updated after each completed phase.
- Keep [DETAILED_BACKLOG_P3_STAGE2.md](DETAILED_BACKLOG_P3_STAGE2.md) in sync with execution status for Stage 1 P3 and Stage 2 P1/P2/P3.
- Keep [TASK_BACKLOG_V21.md](TASK_BACKLOG_V21.md) in sync with v2.1 execution status.

