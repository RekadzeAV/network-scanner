# D-Track Evidence Template

Шаблон для фиксации результатов проверки hardening-задач D-трека в PR и релизных заметках.

## D-Track Evidence

### 1) Smoke / CI

- D-track smoke script:
  - [ ] `./scripts/smoke-d-track-topology-export.sh` (Linux/macOS)
  - [ ] `.\scripts\smoke-d-track-topology-export.ps1` (Windows)
- CI job:
  - [ ] `D-Track Smoke (Topology Export)` — PASS
- Evidence:
  - [ ] URL CI run:
  - [ ] Краткий лог/вывод ключевых шагов:

### 2) Export Consistency (`json` vs `graphml`)

- [ ] Набор узлов эквивалентен между форматами
- [ ] Набор связей эквивалентен между форматами
- [ ] GraphML содержит ключевые поля edge metadata:
  - `source_type`
  - `confidence`
  - `evidence`

### 3) Graphviz Fallback (`png`/`svg`)

- [ ] При наличии `dot` создается целевой файл (`png`/`svg`)
- [ ] При отсутствии `dot`:
  - [ ] команда не падает;
  - [ ] создается fallback `json`;
  - [ ] есть диагностическое сообщение в выводе

### 4) External GraphML Compatibility

- [ ] yEd import — PASS
- [ ] Gephi import — PASS
- [ ] Версии инструментов зафиксированы:
  - yEd:
  - Gephi:
- [ ] Проверка выполнялась по:
  - [GRAPHML_COMPATIBILITY_CHECK.md](GRAPHML_COMPATIBILITY_CHECK.md)

### 5) Notes / Residual Risk

- Известные ограничения:
- Что проверять дополнительно после merge:
