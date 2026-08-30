# Prompt Execution Roadmap

## Current Status

### Phase 1

- 1.1 GUI/Scanner separation: **done**
  - daemon lifecycle hardened (`single-run`, `error` events, state cleanup)
  - GUI scan orchestration isolated to runner bridge/controller path
  - CLI sync path preserved
- 1.2 Interactive topology: **planned**
- 1.3 Large result optimization: **in progress**
  - results render debounce для фильтров/сортировки
  - virtualized cards (`widget.List`) + progressive load (`Показать еще`)
  - host details cache + nearby prefetch
  - benchmark baseline для пайплайна фильтрации/сортировки

### Phase 2

- 2.1 Live inventory (SQLite): **planned**
- 2.2 Device profiling enhancement: **planned**
- 2.3 Human-centric security dashboard: **planned**

### Phase 3

- 3.1 Simple Mode GUI: **planned**
- 3.2 Installers packaging: **planned**
- 3.3 CLI JSON/daemon mode: **planned**

## Milestones

1. M1: close phase 1 (daemon + topology + scalable results).
2. M2: close phase 2 (inventory + classifier + security index).
3. M3: close phase 3 (simple mode + installers + CI daemon mode).

## Validation Gates

- Gate A: `go test ./...` green after each milestone.
- Gate B: GUI smoke scenarios pass (`Scan`, `Stop`, `Topology`, `Security`).
- Gate C: CLI smoke scenarios pass including JSON mode.
