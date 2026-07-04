# Prompt Execution Development Log

## 2026-04-26

### Implemented

- Добавлен новый пакет `internal/scanner/daemon`:
  - `Config` для параметров запуска;
  - `Runner` для lifecycle-управления;
  - событийная модель `Event` (`progress`, `done`, `error`, `stopped`);
  - поддержка неблокирующей доставки событий в GUI.
- GUI (`internal/gui/app.go`) переведен на использование daemon runner:
  - запуск сканирования через runner;
  - обработка прогресса/завершения через `runner.Events()`;
  - совместимость с текущими UI update handlers.
- Обновлен stop path:
  - `stopScan()` теперь останавливает runner и scanner корректно;
  - при completion/timeout обнуляется ссылка на runner.
- Добавлены unit-тесты для daemon пакета:
  - `TestNewRunner`
  - `TestStopWithoutStart`
- Доработан lifecycle `internal/scanner/daemon.Runner`:
  - запрет повторного `Start` во время активного run;
  - возврат ошибки из `Start(...)` при конфликте запуска и `nil` factory result;
  - `EventError` при ошибке и штатный cleanup (`running/cancel/scanner`) после terminal state.
- GUI scan-flow переведен на runner-only orchestration:
  - убрана прямая зависимость GUI от `*scanner.NetworkScanner`;
  - `startScan()` использует `runner.Start(...)` c error-path;
  - event-loop вынесен в `scan_controller.go` (`observeScanRunner(...)`);
  - timeout/stop управляются через `scanRunner.Stop()`.

### File Structure Changes

- `internal/scanner/daemon/daemon.go` (new)
- `internal/scanner/daemon/daemon_test.go` (new)

### Pending (next increments)

- topology interactive graph view package.
- inventory SQLite package and CLI flags.
- GUI simple mode and installers scripts.
- daemon periodic mode in CLI.
