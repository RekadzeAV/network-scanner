# PR Description: UI Results View Improvements

## Summary

- Implemented `Сканирование` result submodes: `Devices` and `Security`.
- `Devices` result submode supports two full result modes: `Таблица` and `Карточки`.
- Added advanced results controls: sorting (`IP`/`HostName`), chip limit (`12/24/48`), text filter, quick type filters, and `Только с открытыми портами`.
- Added filter UX improvements: active-filter counter, quick clear for search input, and full filter reset action.
- Made results settings persistent across app restarts (mode, sorting, chip limit, filters).
- Updated export behavior: `Сохранить результаты` now exports the **currently displayed** subset (after filters + sorting).
- Added `Host Details Drawer` in `Devices` result submode:
  - selected host details (`Host/IP/MAC/Type/Vendor/OS/SNMP/Open ports`);
  - quick actions (`Ping`/`Traceroute`/`DNS`/`Whois`/`Wake-on-LAN`).
- Added `Security Dashboard` in `Security` result submode:
  - aggregated findings (`audit + risk signatures`);
  - findings table (`source/severity/host/title`);
  - `Export security report (HTML)` action.
- Added GUI scan auto-profile controls and UX:
  - `Автопрофиль сканирования (рекомендуется)` toggle;
  - info button (`Почему изменены параметры?`) with threshold logic;
  - visual state indicators (`ВКЛ/ВЫКЛ`) in scan controls and results header.
- Stabilized GUI tools output (`Ping`/`Traceroute`) by using markdown-compatible raw output section (no HTML `<details>`).
- Added RU parsing support for Windows `ping` statistics.
- Added `Operations Runtime` + `Operations Center` in `Инструменты`:
  - lifecycle statuses (`queued/running/success/failed/canceled`);
  - operation history;
  - action controls `Retry`/`Cancel` where applicable.
- Improved scanning performance guardrails:
  - faster host liveness probe behavior;
  - adaptive per-host port probe concurrency limit.
- Improved maintainability by splitting large GUI logic into dedicated files:
  - `internal/gui/scan_controller.go`
  - `internal/gui/topology_controller.go`
  - `internal/gui/operations.go`
  - `internal/gui/security_view.go`
  - `internal/gui/results_view.go`
  - `internal/gui/results_analytics_view.go`
  - `internal/gui/results_charts.go`
  - `internal/gui/results_model.go`
- Added/updated tests around sorting/filtering/model logic and synced user-facing docs/checklists/changelog.

## Test plan

- [ ] Build and run GUI:
  - `go build -o network-scanner-gui ./cmd/gui`
  - Launch app and open `Сканирование`.
- [ ] Run scan and verify result states:
  - scanning state, completed state, stop state, timeout state.
- [ ] Verify result submode switching:
  - `Devices` and `Security` both render correctly.
- [ ] Verify `Devices` + `Таблица` mode:
  - columns `HostName`, `IP`, `MAC`, `Порты`;
  - chips wrap; horizontal scroll works on narrow width;
  - markdown analytics summary visible.
- [ ] Verify `Devices` + `Карточки` mode:
  - cards include HostName/IP/MAC/chips;
  - responsive layout collapses to one column on narrow width;
  - two pie charts with legends/percentages rendered.
- [ ] Verify `Host Details Drawer`:
  - host selection from table/cards updates drawer;
  - host details are shown correctly;
  - quick actions trigger corresponding tools with prefilled host data.
- [ ] Verify `Security` result submode:
  - security summary and findings table are rendered;
  - `Export security report (HTML)` works without errors.
- [ ] Verify controls:
  - sorting (`IP`/`HostName`);
  - chip limit (`12`/`24`/`48`);
  - text filter;
  - quick type filters and `Только с открытыми портами`;
  - active-filter counter updates;
  - `Очистить` and `Сбросить фильтры` actions.
- [ ] Verify scan auto-profile UX:
  - toggle on/off changes indicator (`Автопрофиль: ВКЛ/ВЫКЛ`);
  - info dialog opens and explains thresholds/limits;
  - auto-profile state is restored after app restart.
- [ ] Verify persistence:
  - restart app; confirm result submode/view/sort/chip-limit/filter settings are restored.
- [ ] Verify export behavior:
  - apply filters, click `Сохранить результаты`, ensure exported file matches currently displayed subset.
- [ ] Regression checks:
  - topology tab still works,
  - PNG preview and topology save still work.
- [ ] Verify `Инструменты` tab regressions:
  - `Ping` and `Traceroute` raw output is readable in GUI,
  - Windows-localized ping output still produces parsed summary values.
- [ ] Verify `Operations Center` in `Инструменты`:
  - operation history updates after tool runs;
  - `Retry` available for `failed/canceled`;
  - `Cancel` available for `queued/running`.
- [ ] Automated checks:
  - `go test ./...`

## D-Track evidence (optional section)

Если изменения затрагивают topology/export hardening, добавьте блок:

```md
## D-Track Evidence

### Smoke / CI
- [ ] D-track smoke script PASS
- [ ] CI job `D-Track Smoke (Topology Export)` PASS
- [ ] URL CI run: <link>

### Export Consistency
- [ ] `json`/`graphml` node-set equivalent
- [ ] `json`/`graphml` edge-set equivalent
- [ ] `source_type`/`confidence`/`evidence` present in GraphML

### Graphviz Fallback
- [ ] `png/svg` export works with `dot`
- [ ] fallback JSON works without `dot`

### External Compatibility
- [ ] yEd import PASS
- [ ] Gephi import PASS
```

Полная версия шаблона: [D_TRACK_EVIDENCE_TEMPLATE.md](D_TRACK_EVIDENCE_TEMPLATE.md).

Готовый короткий вариант для вставки: [D_TRACK_EVIDENCE_PR_SNIPPET.md](D_TRACK_EVIDENCE_PR_SNIPPET.md).

## D-Track Evidence (current snapshot)

### Smoke / CI
- [x] Windows smoke: `.\scripts\smoke-d-track-topology-export.ps1` PASS
- [ ] Linux/macOS smoke: `./scripts/smoke-d-track-topology-export.sh` (pending)
- [ ] CI job `D-Track Smoke (Topology Export)` (pending)
- [ ] CI URL: pending

### Export Consistency (`json` vs `graphml`)
- [x] Node-set equivalence: PASS
- [x] Edge-set equivalence: PASS
- [x] GraphML metadata keys present: `source_type`, `confidence`, `evidence`

### Graphviz Fallback
- [x] Without `dot`: command does not fail, fallback JSON is generated, diagnostic message is present
- [ ] With `dot`: direct `png/svg` generation (pending)

### External Compatibility
- [ ] yEd import PASS (pending)
- [ ] Gephi import PASS (pending)

### Residual Risk
- Pending: external imports (yEd/Gephi), CI confirmation, and `dot`-available validation.
