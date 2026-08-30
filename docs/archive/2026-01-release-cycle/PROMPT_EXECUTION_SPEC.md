# Prompt Execution Specification

## Goal

Последовательно реализовать 3-фазный план развития network-scanner без поломки текущей архитектуры и с сохранением совместимости CLI/GUI.

## Scope Breakdown

### Phase 1 (stability and architecture debt)

1. Разделение GUI и engine через `internal/scanner/daemon`.
2. Развитие topology UI до интерактивной карты.
3. Оптимизация рендеринга результатов для больших сетей.

### Phase 2 (feature deepening)

1. Live inventory в SQLite: snapshot + diff.
2. Расширенная классификация устройств.
3. Human-readable security findings и network security index.

### Phase 3 (positioning and growth)

1. GUI Simple/Advanced mode.
2. Installer packaging scripts (Windows/macOS/Linux).
3. CLI JSON-only output contract + periodic daemon mode.

## Non-Functional Requirements

- Новый код только на Go.
- GUI только на Fyne v2.
- Совместимость с текущими internal-пакетами и CLI интерфейсом.
- Поддержка `go test ./...` после каждого этапа.

## API Compatibility Notes

- `internal/scanner` остается источником синхронного scanning API.
- `internal/scanner/daemon` вводится как дополнительный orchestration слой.
- Публичные сигнатуры существующих пакетов не ломаются.

## Deliverables Tracking

- [x] Фундамент daemon-слоя (phase 1.1).
- [x] Полное закрытие phase 1.1 (daemon lifecycle hardening + GUI runner-only orchestration).
- [ ] Интерактивная topology map.
- [ ] Large-results optimization.
- [ ] Inventory subsystem.
- [ ] Classifier overhaul.
- [ ] Security index + human-readable findings.
- [ ] Simple Mode.
- [ ] Installers scripts.
- [ ] CLI daemon schedule + JSON contract.
