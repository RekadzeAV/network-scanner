# Commit Ready: Stage2 Sign-off Pack

Готовые варианты commit message для текущего пакета изменений.

## Вариант 1 (рекомендуется)

```text
docs: finalize Stage2 sign-off workflow and sync release status

Align Stage1/Stage2 phase status to 100% in roadmap/checklists, add unified
P0 preflight+runbook flow, and fix Stage2 P3 closure script false-fail behavior
on PowerShell stderr warnings.
```

## Вариант 2 (короче)

```text
docs+scripts: complete Stage2 closure docs and P0 preflight flow

Sync release/closure documentation, add blocker diagnostics and runbooks,
and stabilize stage2-p3 PowerShell closure checks.
```

## Вариант 3 (если в одном коммите и GUI, и доки)

```text
release: Stage2 sign-off pack, GUI scan layout, and closure tooling

Document 100% Stage1/Stage2 phase closure, add preflight/docs-link/status
scripts, fix stage2-p3 PS stderr handling, improve scan tab VSplit and
adaptive layout scale tracking; ignore local build/ in git.
```

## Что входит в пакет

- Синхронизация статусов Stage 1/Stage 2 в:
  - roadmap/backlog/checklists/snapshot
- Единый operational контур sign-off:
  - `RELEASE_READY_GAP_LIST`
  - `CHECKLIST_STATUS_INDEX`
  - `P0_SIGNOFF_RUNBOOK`
  - `RELEASE_OPERATIONS_CHEATSHEET`
  - `STAGE2_100_COMMIT_READY`
- Автопроверка блокеров:
  - `scripts/p0-signoff-preflight.ps1`
  - `Makefile` цель `p0-preflight-win`
- Доп. sanity перед коммитом:
  - `scripts/docs-link-check.ps1` + `docs-link-check-win`
  - `scripts/stage2-signoff-status.ps1` + `stage2-signoff-status-win` (агрегат: Stage2 P1–P3 + ссылки в docs + P0 preflight)
- Исправление closure-script стабильности:
  - `scripts/stage2-p3-closure-check.ps1`
- GUI (вкладка `Сканирование` / адаптив):
  - `internal/gui/app.go` — `VSplit` скан/результаты, scroll для tools/operations
  - `internal/gui/results_view.go` — согласованные минимальные размеры и обновление `lastCanvasScale` при смене профиля layout
- Репозиторий:
  - `.gitignore` — `build/` (в т.ч. `build/release/` для выхода релизных скриптов), корневой `release/` для старых локальных раскладок, типичные `security-report-{redacted,unredacted}-*.html` в корне после ручных CLI-прогонов
