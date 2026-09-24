# Справка по рефакторингу документации (2026-09-22)

Данный файл фиксирует что было проанализировано, что объединено, какие ссылки исправлены, что перемещено в архив и почему.

## Что проанализировано
- Все действующие `docs/*.md`
- Все `docs/archive/**`
- `internal/legacy/*`, `internal/zaclikivaniya.md`, `internal/benchmark/benchmarks_baseline.txt`
- Корневые scratch-файлы (`pr1.txt`, `pr2.txt`, `pr3.txt`, `rf.txt`, `rs.txt`, `s2b.txt`, `s3.txt`, `sc_check.txt`, `sk.txt`, `sk2.txt`, `tk.txt`, `cover_func.txt`)
- `cmd/gui/BUILD_WINDOWS.md`, `Инструкция по эксплуатации.md`, `QUICKSTART_WINDOWS_BUILD.md`

## Что объединено / выявлено как однотипное
- Несколько индексных/навигационных файлов дублировали друг друга. Индекс навигации теперь ведётся из `docs/README.md` и `README.md` (правило записано в `docs/ROADMAP.md`).
- Планы циклов (FINAL_PLAN / UNIFIED_OPTIMIZED_PLAN / THREE_PLANS_ANALYSIS / DETAILED_PLAN_V22 / BLOCKER / DOCUMENTATION_UPDATE_REPORT / D_TRACK / GUI_SMOKE_CHECKLIST) имели перекрывающиеся секции статусов. Фактический статус сейчас сведён в `docs/IMPLEMENTATION_PLAN.md` + `docs/ROADMAP.md` + `docs/CHANGELOG.md` + `CHANGELOG.md`.

## Перемещено в архив (итог — только завершённые артефакты циклов)
- `docs/TASK_BACKLOG_V21.md` → `docs/archive/2026-09-audit/TASK_BACKLOG_V21.md`
- `docs/INSTALL_OLD.md` → `docs/archive/2026-09-audit/INSTALL_OLD.md`
- `docs/RELEASE_MANIFEST_V2.md` → `docs/archive/2026-09-audit/RELEASE_MANIFEST_V2.md`
- `docs/RELEASE_INSTRUCTIONS_V22.md` → `docs/archive/2026-09-audit/RELEASE_INSTRUCTIONS_V22.md`
- `docs/RELEASE_NOTES_V22.md` → `docs/archive/2026-09-audit/RELEASE_NOTES_V22.md`
- `docs/FINAL_RELEASE_READINESS_REPORT.md` → `docs/archive/2026-09-audit/FINAL_RELEASE_READINESS_REPORT.md`
- `docs/GRAPHML_COMPATIBILITY_CHECK.md` — остаётся активным (ссылается `docs/README.md`)
- `docs/GIT_SETUP.md` — остаётся активным (ссылается `docs/README.md`)
- `docs/BLOCKER_ZACILIVANIE_001.md` → `docs/archive/2026-09-audit/BLOCKER_ZACILIVANIE_001.md`
- `docs/DETAILED_PLAN_V22.md` → `docs/archive/2026-09-audit/DETAILED_PLAN_V22.md`
- `docs/COVERAGE_STATUS.md` → `docs/archive/2026-09-15-docs-sync/COVERAGE_STATUS.md`
- `docs/DOCUMENTATION_UPDATE_REPORT_2026-08-18.md` → `docs/archive/2026-09-15-docs-sync/DOCUMENTATION_UPDATE_REPORT_2026-08-18.md`
- `Инструкция по эксплуатации.md` (корень, untracked, v1.0.x) → `docs/archive/2026-09-22-docs-cleanup/`
- `internal/zaclikivaniya.md` → `docs/archive/2026-09-15-docs-sync/zaclikivaniya-notes.md`
- Корневые scratch-файлы (`pr1.txt`, `pr2.txt`, `pr3.txt`, `rf.txt`, `rs.txt`, `s2b.txt`, `s3.txt`, `sc_check.txt`, `sk.txt`, `sk2.txt`, `tk.txt`, `cover_func.txt`) — удалены из репозитория, аналогичного мусора быть не должно (см. `.gitignore` и F4/F5 в `FINAL_PLAN_2026-09-15.md`).

## Возвращено из архива в `docs/` (активная документация)
Первичная версия рефакторинга перенесла в архив документы, которые активный индекс
`docs/README.md` и корневой `README.md` объявляют текущими, — из-за чего возникли
битые ссылки (критерий завершения E2 «нет битых ссылок» нарушался). Возвращены без
изменения содержимого:

- продуктовая документация: `ARCHITECTURE.md`, `TECHNICAL.md`, `USER_GUIDE.md`, `GUI.md`, `GUI_SMOKE_CHECKLIST.md`, `TEST_NETWORK_PROFILE.md`, `GRAPHML_COMPATIBILITY_CHECK.md`, `LOGGING.md`, `deployment.md`
- установка/сборка: `INSTALL_WINDOWS.md`, `INSTALL_LINUX_CROSS_COMPILER.md`, `SETUP_WINDOWS_CROSS_COMPILE.md`, `CROSS_COMPILATION_WINDOWS.md`, `CROSS_COMPILATION_QUICKREF.md`, `BUILD_REQUIREMENTS_WINDOWS.md`, `BUILD_STRUCTURE.md`, `QUICKSTART-macOS.md`, `GIT_SETUP.md`

Их дубли-копии, созданные в `docs/archive/2026-09-15-docs-sync/` и
`docs/archive/2026-09-audit/`, удалены — дублей активной документации быть не должно.

## Исправленные ссылки
- `docs/README.md`: убрано указание на заменённый файл `UNIFIED_OPTIMIZED_PLAN.md` (2026-09-04) — осталось на актуальный `UNIFIED_OPTIMIZED_PLAN_2026-09-15.md`.
- `docs/IMPLEMENTATION_PLAN.md`: исправлена справка «Карта покрытия трёх планов» — ссылка на `UNIFIED_OPTIMIZED_PLAN.md` заменена на `UNIFIED_OPTIMIZED_PLAN_2026-09-15.md`.
- `docs/INSTALL.md`: `TASK_BACKLOG_V21.md` переведён на архивный путь, `ROADMAP.md` — на `docs/ROADMAP.md`.
- `docs/ARCHITECTURE.md`: ссылка на перемещённую в архив `Инструкция по эксплуатации.md` обновлена.
- `scripts/final-release-check.{sh,ps1}`: docs-sanity больше не требует отчёты
  завершённого цикла (`FINAL_RELEASE_READINESS_REPORT.md`,
  `D_TRACK_IMPLEMENTATION_STATUS.md`) — проверяются активные документы.

## Инструмент проверки ссылок
`scripts/docs-link-check.ps1` по умолчанию исключает `docs/archive/**`
(исторические ссылки завершённых циклов не восстанавливаются). Полный обход
архива — с флагом `-IncludeArchive`.


## Не трогалось (обоснование)
- `internal/legacy/*`, `internal/benchmark/*`, `Vendor list/oui.txt`, `Update-project/*`, `cmd/gui/BUILD_WINDOWS.md`, `docs/adr/*`, `docs/archive/*` — исторические/релизные/legacy-артефакты, не являются дублями активной навигации.
- `QUICKSTART_WINDOWS_BUILD.md` — активен (на него ссылается корневой `README.md`), ссылки внутри целы.

## Проверка после изменений
- `.\scripts\docs-link-check.ps1` — OK: битых локальных ссылок в активной документации нет.
- `docs/archive/**` из проверки исключён (см. «Инструмент проверки ссылок»): там
  остаются исторические ссылки завершённых циклов, восстановлению не подлежат.
- Дубли активного содержимого в `docs/archive/**` удалены.

## Что ещё в рабочем дереве
- `docs/DOCS_REFACTOR_NOTES.md` — этот файл.
- `docs/PROJECT_PROMPT.md` — новый файл (единый промт проекта, вне этого рефакторинга).
- `docs/archive/2026-09-15-docs-sync/COVERAGE_STATUS.md`,
  `docs/archive/2026-09-15-docs-sync/DOCUMENTATION_UPDATE_REPORT_2026-08-18.md` —
  новые архивные копии перемещённых отчётов.

