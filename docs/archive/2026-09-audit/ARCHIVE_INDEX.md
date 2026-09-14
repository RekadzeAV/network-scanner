# Archive Index

**Дата создания:** 2026-09-02  
**Цель:** Индекс всех архивных документов проекта

---

## Структура архива

```
docs/archive/
├── 2026-01-release-cycle/      # Релизный цикл v2.0-v2.1 (31 файл)
├── 2026-04-docs-sync/          # Синхронизация документации (5 файлов)
├── 2026-08-ui-tests/           # Тесты UI результатов (3 файла)
└── 2026-09-audit/              # Аудит и оптимизация (3 файла)
```

---

## 2026-09-audit/ — Аудит и оптимизация

| Файл | Описание | Дата |
|------|----------|------|
| `AUDIT_REPORT.md` | Полный отчёт аудита документации vs код | 2026-09-02 |
| `OPTIMIZATION_PLAN.md` | План оптимизации и рефакторинга | 2026-09-02 |
| `SUMMARY.md` | Итоговый краткий отчёт | 2026-09-02 |

### Ключевые выводы:

- **108 файлов** в internal/gui/ (было ~30)
- **app.go:** 2780 → 605 строк (-78%)
- **IPv6 Support** реализован, но не задокументирован
- **11 устаревших файлов** перемещено в архив
- **23 часа** работы по оптимизации запланировано

---

## 2026-01-release-cycle/ — Релизный цикл v2.0-v2.1

| Файл | Описание | Статус |
|------|----------|--------|
| `PLAN_STABILIZATION_V2.md` | План стабилизации | ✅ Архив |
| `ROLLBACK_PLAN_D_TRACK.md` | План отката D-Track | ✅ Архив |
| `DETAILED_IMPLEMENTATION_PLAN_v2.md` | Детальный план v2 | ✅ Архив |
| `DOCUMENTATION_UPDATE_REPORT_2026-01-XX.md` | Отчёт обновления | ✅ Архив |
| `DEVELOPMENT_PLAN_v1.1.md` | План v1.1 | ✅ Архив |
| `STAGE2_100_COMMIT_READY.md` | Stage2 commit ready | ✅ Архив |
| `RELEASE_SUMMARY_STAGE2_P3.md` | Summary P3 | ✅ Архив |
| `RELEASE_SUMMARY_STAGE2_P2.md` | Summary P2 | ✅ Архив |
| `RELEASE_READINESS_STAGE2_P3_PR_SNIPPET.md` | Readiness P3 | ✅ Архив |
| `RELEASE_READINESS_PR_SNIPPET.md` | Readiness PR | ✅ Архив |
| `RELEASE_READINESS_PR_READY.md` | PR ready | ✅ Архив |
| `PROMPT_EXECUTION_SPEC.md` | Execution spec | ✅ Архив |
| `PROMPT_EXECUTION_ROADMAP.md` | Execution roadmap | ✅ Архив |
| `PROMPT_EXECUTION_DEVELOPMENT_LOG.md` | Development log | ✅ Архив |
| `PROMPT_EXECUTION_ANALYSIS.md` | Analysis | ✅ Архив |
| `PLATFORM_DECOMPOSITION.md` | Platform decomp | ✅ Архив |
| `PLAN_IMPLEMENTATION_v1.1.md` | Plan v1.1 | ✅ Архив |
| `D_TRACK_EVIDENCE_TEMPLATE.md` | Evidence template | ✅ Архив |
| `D_TRACK_EVIDENCE_PR_SNIPPET.md` | Evidence PR | ✅ Архив |
| `D_TRACK_EVIDENCE_CURRENT.md` | Evidence current | ✅ Архив |
| `MANUAL_SIGNOFF_DRAFT.md` | Signoff draft | ✅ Архив |
| `COMMIT_READY_STAGE2_SIGNOFF.md` | Signoff ready | ✅ Архив |
| `MANUAL_SIGNOFF_TEMPLATE.md` | Signoff template | ✅ Архив |
| `PHASE_1_1_CLOSURE_CHECKLIST.md` | Closure checklist | ✅ Архив |
| `PHASE_1_1_AUTOMATION_REPORT.md` | Automation report | ✅ Архив |
| `PHASE2_TASKS_CHECKLIST.md` | Phase2 tasks | ✅ Архив |
| `P3_PERF_BASELINE.md` | Perf baseline | ✅ Архив |
| `P3_CLOSURE_CHECKLIST.md` | P3 closure | ✅ Архив |
| `P1_CLOSURE_CHECKLIST.md` | P1 closure | ✅ Архив |
| `RELEASE_NOTES_V22.md` | Release notes v22 | ✅ Архив |
| `DETAILED_PLAN_V22.md` | Detailed plan v22 | ✅ Архив |

---

## 2026-04-docs-sync/ — Синхронизация документации

| Файл | Описание | Статус |
|------|----------|--------|
| `RELEASE_READY_GAP_LIST.md` | Gap list | ✅ Архив |
| `FINAL_PR_COMMENT_STAGE2_P3_READY.md` | PR comment P3 | ✅ Архив |
| `FINAL_PR_COMMENT_READY.md` | PR comment | ✅ Архив |
| `DOCS_SYNC_SUMMARY_2026-04-23.md` | Summary | ✅ Архив |
| `DOCS_SYNC_PR_SNIPPET_2026-04-23.md` | PR snippet | ✅ Архив |
| `DOCS_SYNC_PR_SNIPPET_2026-04-23_EN.md` | PR snippet EN | ✅ Архив |

---

## 2026-08-ui-tests/ — Тесты UI результатов

| Файл | Описание | Статус |
|------|----------|--------|
| `RELEASE_SUMMARY_UI_RESULTS.md` | Summary UI | ✅ Архив |
| `PR_DESCRIPTION_UI_RESULTS.md` | PR description | ✅ Архив |
| `DECOMPOSITION_H1_GUI_TESTS.md` | Decomposition | ✅ Архив |

---

## Правила архивирования

### Когда перемещать:

1. **Устаревшие версии планов** — когда есть актуальная версия
2. **Закрытые blockers** — когда задача выполнена
3. **Отчёты аудита** — после проведения следующего аудита
4. **Release notes** — когда выпущена новая версия
5. **Backlog** — когда есть актуальный backlog

### Как архивировать:

1. Создать папку `docs/archive/YYYY-MM-name/`
2. Переместить файлы с датой < 2 месяцев
3. Обновить `ARCHIVE_INDEX.md`
4. Удалить ссылки на архивные файлы из актуальной документации

### Что НЕ архивировать:

- `README.md`
- `CHANGELOG.md`
- `LICENSE`
- Актуальная документация (ROADMAP, ARCHITECTURE, IMPLEMENTATION_PLAN)
- Тесты и код

---

**Индекс создан:** 2026-09-02  
**Последнее обновление:** 2026-09-02  
**Статус:** ✅ Актуален
