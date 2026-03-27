# PR Description: UI Results View Improvements

## Summary

- Implemented two full GUI result modes in `Сканирование`: `Таблица` and `Карточки`.
- Added advanced results controls: sorting (`IP`/`HostName`), chip limit (`12/24/48`), text filter, quick type filters, and `Только с открытыми портами`.
- Added filter UX improvements: active-filter counter, quick clear for search input, and full filter reset action.
- Made results settings persistent across app restarts (mode, sorting, chip limit, filters).
- Updated export behavior: `Сохранить результаты` now exports the **currently displayed** subset (after filters + sorting).
- Improved maintainability by splitting large GUI logic into dedicated files:
  - `internal/gui/results_view.go`
  - `internal/gui/results_charts.go`
  - `internal/gui/results_model.go`
- Added/updated tests around sorting/filtering/model logic and synced user-facing docs/checklists/changelog.

## Test plan

- [ ] Build and run GUI:
  - `go build -o network-scanner-gui ./cmd/gui`
  - Launch app and open `Сканирование`.
- [ ] Run scan and verify result states:
  - scanning state, completed state, stop state, timeout state.
- [ ] Verify mode switching:
  - `Таблица` and `Карточки` both render correctly.
- [ ] Verify `Таблица` mode:
  - columns `HostName`, `IP`, `MAC`, `Порты`;
  - chips wrap; horizontal scroll works on narrow width;
  - protocol/device analytics blocks visible.
- [ ] Verify `Карточки` mode:
  - cards include HostName/IP/MAC/chips;
  - responsive layout collapses to one column on narrow width;
  - two pie charts with legends/percentages rendered.
- [ ] Verify controls:
  - sorting (`IP`/`HostName`);
  - chip limit (`12`/`24`/`48`);
  - text filter;
  - quick type filters and `Только с открытыми портами`;
  - active-filter counter updates;
  - `Очистить` and `Сбросить фильтры` actions.
- [ ] Verify persistence:
  - restart app; confirm view/sort/chip-limit/filter settings are restored.
- [ ] Verify export behavior:
  - apply filters, click `Сохранить результаты`, ensure exported file matches currently displayed subset.
- [ ] Regression checks:
  - topology tab still works,
  - PNG preview and topology save still work.
- [ ] Automated checks:
  - `go test ./...`
