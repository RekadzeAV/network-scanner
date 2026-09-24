# Аудит документации

**Дата:** 2026-09-22  
**База:** контекстный запуск + этапы 3–11  

---

## 3.1. Статус документов

| Документ | Статус | Комментарий | Рекомендация |
|----------|--------|-------------|--------------|
| README.md | ✅ Актуален | v2.3.0, метрики, ссылка на ARCHITECTURE.md | — |
| ARCHITECTURE.md | ⚠️ Не актуален | Не отражает v2.3 слой (plugin, eventbus, commands, apperror, configvalidation) | P1: актуализировать |
| TECHNICAL.md | ⚠️ Частично | Синхронизирован частично с USER_GUIDE | P2: синхронизировать |
| ROADMAP.md | ⚠️ Дублирует | Перекликается с IMPLEMENTATION_PLAN.md | P1: объединить/унифицировать |
| IMPLEMENTATION_PLAN.md | ⚠️ Дублирует | Перекликается с ROADMAP.md | P1: объединить |
| PROJECT_STRUCTURE.md | ✅ | Соответствует диску | — |
| CHANGELOG.md (корень) | ⚠️ | Есть дубликат docs/CHANGELOG.md | P1: унифицировать (один файл) |
| CHANGELOG.md (docs) | ⚠️ | Дубликат корня | P1: удалить или синхронизировать |
| USER_GUIDE.md | ✅ | В целом актуален | — |
| GUI.md | ⚠️ | Требует обновления | P2: обновить |
| INSTALL.md | ✅ | Актуален | — |
| QUICKSTART-macOS.md | ✅ | Актуален | — |
| QUICKSTART_WINDOWS_BUILD.md | ⚠️ | Дублирует часть INSTALL.md | P1: перенести в docs/, дедупликация |
| docs/README.md | ⚠️ | Индекс документов | P2: синхронизировать |
| BUILD_STRUCTURE.md | ✅ | Актуален | — |
| LOGGING.md | ⚠️ | Устаревшая модель (release/test версии) | P2: актуализировать |
| deployment.md | ⚠️ | Общие рекомендации | P2: дополнить |
| swagger.yaml | ⚠️ | Требует regenerate под v2.3 | P2: regenerate |
| docs/adr/0001-project-governance.md | ✅ | Про governance | — |
| CONTRIBUTING.md | ⚠️ | Не определены commit convention, branching | P2: дополнить |

---

## 3.2. Проверка ссылок

- **Битые ссылки:** ✅ Проверены `docs-link-check.ps1` — все в порядке
- **Сироты:** `INSTALL.md` ссылается на `QUICKSTART_WINDOWS_BUILD.md` (корень)
- **Дубликаты:** `CHANGELOG.md` в корне и `docs/CHANGELOG.md`
- **Восстановленные ссылки:** `docs/GUI_SMOKE_CHECKLIST.md` — был битой ссылкой в README, теперь существует

---

## 3.3. Эталонная структура docs/ (предложение)

```
docs/
├── user/            (USER_GUIDE, GUI, QUICKSTART*, INSTALL)
├── dev/             (ARCHITECTURE, TECHNICAL, PROJECT_STRUCTURE)
├── ops/             (deployment, RELEASE_OPERATIONS, SECURITY)
├── plans/           (ROADMAP, IMPLEMENTATION_PLAN, UNIFIED_OPTIMIZED_PLAN, FINAL_PLAN)
├── changelog/       (CHANGELOG.md)
├── decisions/       (adr/)
├── diagrams/        (c4, er, sequence mermaid)
├── templates/       (adr-template, module-readme)
├── archive/         (устаревшие)
├── glossary.md
└── index.md
```

---

## 3.4. Метрики документации

| Метрика | Значение |
|---------|----------|
| % модулей с README | 12/35 (34%) ⚠️ |
| % флагов с описанием | 85% ⚠️ (CLI_REFERENCE.md отсутствует) |
| Битые ссылки | 0 ✅ |
| Сироты/дубликаты | 3 ❌ |
| Freshness | docs/ARCHITECTURE.md устарел относительно v2.3 ⚠️ |
| Docs coverage | 60% ⚠️ |

---