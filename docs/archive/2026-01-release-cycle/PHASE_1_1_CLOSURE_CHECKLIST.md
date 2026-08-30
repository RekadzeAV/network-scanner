# Phase 1.1 Closure Checklist

Цель: формально закрыть этап `1.1 — Разделение GUI и движка сканирования`.

## 1) Daemon lifecycle

- [x] `Runner` поддерживает только один активный run.
- [x] Повторный `Start(...)` во время активного run возвращает ошибку.
- [x] `Start(...)` возвращает ошибку при невалидном startup (`nil` scanner factory result).
- [x] `EventError` эмитится при startup/lifecycle ошибках.
- [x] После terminal state очищаются внутренние ссылки (`running/cancel/scanner`).

## 2) GUI/engine boundary

- [x] GUI orchestration выполняется через `scanRunner`.
- [x] Удален прямой control path через `*scanner.NetworkScanner`.
- [x] `startScan()` обрабатывает ошибку `runner.Start(...)`.
- [x] Event-loop scan runner вынесен в контроллер (`observeScanRunner(...)`).
- [x] `stop/timeout` идут через `scanRunner.Stop()`.

## 3) Test coverage

- [x] Unit: базовое создание runner (`TestNewRunner`).
- [x] Unit: stop без запуска (`TestStopWithoutStart`).
- [x] Unit: защита от повторного старта (`TestStartRejectsWhenAlreadyRunning`).
- [x] Unit: error path (`TestStartEmitsErrorWhenFactoryReturnsNil`).

## 4) Compatibility and docs

- [x] CLI sync path сохранен (использует `internal/scanner` напрямую).
- [x] `PROMPT_EXECUTION_ROADMAP.md` обновлен: `1.1 -> done`.
- [x] `PROMPT_EXECUTION_SPEC.md` фиксирует полное закрытие phase 1.1.
- [x] `PROMPT_EXECUTION_DEVELOPMENT_LOG.md` синхронизирован по изменениям.

## 5) Validation

- [x] `go test ./...` полностью green в текущем дереве.
- [x] `scripts/p1-closure-check.ps1` проходит на Windows.
- [x] `scripts/integration-check.ps1` проходит на Windows (CLI smoke chain).
- [ ] GUI manual smoke (states `Scan/Stop/Topology/Security` + resolution matrix) выполнен вручную по `docs/GUI_SMOKE_CHECKLIST.md`.
