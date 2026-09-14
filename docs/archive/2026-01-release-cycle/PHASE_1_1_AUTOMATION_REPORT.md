# Phase 1.1 Automation Report

Дата: `2026-04-26`  
ОС: `Windows amd64`

## Scope

Отчет фиксирует автоматические шаги закрытия этапа `1.1 — Разделение GUI и движка сканирования`.

## Implemented

- Усилен lifecycle `internal/scanner/daemon.Runner`:
  - `Start(...)` возвращает `error`;
  - защита от повторного старта во время активного запуска;
  - `EventError` для error-path;
  - cleanup внутреннего состояния после terminal events;
  - добавлен `IsRunning()`.
- GUI переведен на runner-only orchestration:
  - удален прямой control-path через `*scanner.NetworkScanner`;
  - `startScan()` работает через `runner.Start(...)` и обрабатывает startup error;
  - stop/timeout идут через `scanRunner.Stop()`;
  - event-loop вынесен в controller (`observeScanRunner(...)`).
- Добавлены/расширены unit-тесты daemon:
  - `TestNewRunner`
  - `TestStopWithoutStart`
  - `TestStartRejectsWhenAlreadyRunning`
  - `TestStartEmitsErrorWhenFactoryReturnsNil`
- Обновлены phase-документы:
  - `PROMPT_EXECUTION_ROADMAP.md` (`1.1 -> done`);
  - `PROMPT_EXECUTION_SPEC.md` (фикс полного закрытия 1.1);
  - `PROMPT_EXECUTION_DEVELOPMENT_LOG.md`.

## Automated Validation Evidence

- [x] `go test ./...`
- [x] `go build -o network-scanner ./cmd/network-scanner`
- [x] `go build -o network-scanner-gui ./cmd/gui`
- [x] `scripts/smoke-cli-no-topology.ps1`
- [x] `scripts/smoke-cli-topology.ps1`
- [x] `scripts/smoke-cli-tools.ps1`
- [x] `scripts/p1-closure-check.ps1`
- [x] `scripts/integration-check.ps1`

## Remaining Manual Gate

- [ ] Выполнить ручной GUI smoke по `docs/GUI_SMOKE_CHECKLIST.md`:
  - состояния `Scan/Stop/Topology/Security`;
  - матрица разрешений из `scripts/smoke-gui-resolution.ps1`.

Только после этого этап `1.1` считается полностью закрытым по UX-гейту.
