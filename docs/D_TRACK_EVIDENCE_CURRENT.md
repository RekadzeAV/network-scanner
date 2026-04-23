# D-Track Evidence (Current Branch Snapshot)

Актуальный снимок проверок для topology/export hardening в текущей ветке.

## D-Track Evidence

### 1) Smoke / CI

- D-track smoke script:
  - [ ] `./scripts/smoke-d-track-topology-export.sh` (Linux/macOS) — pending (локально не запускался в текущем Windows окружении)
  - [x] `.\scripts\smoke-d-track-topology-export.ps1` (Windows) — PASS
- CI job:
  - [ ] `D-Track Smoke (Topology Export)` — pending (ожидает run в GitHub Actions)
- Evidence:
  - [ ] URL CI run: pending
  - [x] Локальный Windows smoke лог: `Smoke passed: D-track topology exports are healthy.`

### 2) Export Consistency (`json` vs `graphml`)

- [x] Набор узлов эквивалентен между форматами (unit + smoke checks)
- [x] Набор связей эквивалентен между форматами (unit + smoke checks)
- [x] GraphML содержит ключевые поля edge metadata:
  - `source_type`
  - `confidence`
  - `evidence`

### 3) Graphviz Fallback (`png`/`svg`)

- [ ] При наличии `dot` создается целевой файл (`png`/`svg`) — pending (нужен run в среде с установленным Graphviz)
- [x] При отсутствии `dot`:
  - [x] команда не падает;
  - [x] создается fallback `json`;
  - [x] есть диагностическое сообщение в выводе

### 4) External GraphML Compatibility

- [ ] yEd import — pending
- [ ] Gephi import — pending
- [ ] Версии инструментов зафиксированы:
  - yEd: pending
  - Gephi: pending
- [x] Проверка должна выполняться по:
  - `docs/GRAPHML_COMPATIBILITY_CHECK.md`

### 5) Notes / Residual Risk

- Известные ограничения:
  - проверки yEd/Gephi и CI-запусков пока не подтверждены в текущей ветке;
  - проверка режима `png/svg` с установленным `dot` не подтверждена локально.
- Что проверить дополнительно после merge:
  - прогон Unix smoke (`.sh`) в CI runner;
  - ручной импорт `graphml` в yEd и Gephi;
  - прикрепить CI URL и версии внешних инструментов в PR evidence-блок.
