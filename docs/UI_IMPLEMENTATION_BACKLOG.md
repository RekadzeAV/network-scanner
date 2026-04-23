# UI Refactor And Feature Backlog

Детализированный список задач для последовательной реализации UI-плана (`P0 -> P5`).

## P0. Stabilization And Baseline

- [x] `P0.1` Зафиксировать baseline:
  - [x] прогон `go test ./...`
  - [x] прогон GUI smoke checklist
  - [x] зафиксировать результаты в release notes/checklist
- [x] `P0.2` Ввести feature flags:
  - [x] `ui.operations_center`
  - [x] `ui.host_details_drawer`
  - [x] `ui.security_dashboard`
  - [x] `ui.scenario_presets`
- [x] `P0.3` Документировать целевую декомпозицию GUI:
  - [x] App как composition root
  - [x] controllers/services/views responsibilities

## P1. Operations Runtime + First Integrations

- [x] `P1.1` Создать `internal/gui/operations.go`:
  - [x] `OperationType`, `OperationStatus`, `Operation`
  - [x] lifecycle `queued/running/success/failed/canceled`
  - [x] `Run`, `Cancel`, `Retry`, `List`, `Subscribe`
- [x] `P1.2` Покрыть `operations` unit-тестами:
  - [x] статусы и переходы
  - [x] подписки на события
  - [x] cancel/retry semantics
- [x] `P1.3` Интегрировать операции в Tools flow:
  - [x] запуск tool-процедур через runtime
  - [x] статус/ошибки через unified operation result
  - [x] сохранить текущий UX вывод в `toolsOutput`
- [x] `P1.4` Добавить минимальный Operations Center view:
  - [x] список последних операций
  - [x] статус, длительность, ошибка
  - [x] действия `Retry`/`Cancel` (где применимо)

## P2. Tools History UX

- [x] `P2.1` Добавить модель истории запусков инструментов
- [x] `P2.2` Добавить список запусков в UI вкладки `Инструменты`
- [x] `P2.3` Реализовать rerun из истории

## P3. Host Details Drawer

- [x] `P3.1` Модель выбранного устройства и синхронизация с фильтрами
- [x] `P3.2` Drawer с деталями устройства
- [x] `P3.3` Quick actions (`Ping/Traceroute/DNS/Whois/WOL`) через operations runtime

## P4. Security Dashboard MVP

- [x] `P4.1` Подрежимы `Devices/Security` на вкладке сканирования
- [x] `P4.2` Unified findings model (`audit + risksignature`)
- [x] `P4.3` Таблица findings + фильтры + severity summary
- [x] `P4.4` Export security report (HTML)

## P5. Scenario Presets + Cleanup

- [x] `P5.1` Единая модель scenario preset
- [x] `P5.2` UI управления сценариями + предустановки
- [x] `P5.3` Устранить дубли фильтрации (`results_view` vs `results_model`)
  - [x] Убран дублирующийся sort-path в `results_view` (переведен на `sortedResultsForDisplayWithMode`)
  - [x] Унифицирована нормализация device category через `normalizeDeviceTypes`
  - [x] Основной пайплайн фильтрации переведен на `filterResultsForDisplayAdvanced` + дополнительные UI-фильтры
  - [x] Финальная консолидация API фильтрации в одном source of truth (legacy helper-методы удалены из `results_view`)
- [x] `P5.4` Явно встроить analytics в пайплайн рендера результатов
  - [x] Добавлен явный analytics-блок в `renderScanResultsView`
  - [x] `Таблица`: markdown analytics summary
  - [x] `Карточки`: pie charts (`Протоколы`, `Типы устройств`)

## Delivery Order (PR sequence)

1. `PR-1`: `P0` + `P1.1`
2. `PR-2`: `P1.2` + `P1.3`
3. `PR-3`: `P1.4` + стабилизация
4. `PR-4`: `P2`
5. `PR-5`: `P3`
6. `PR-6`: `P4`
7. `PR-7`: `P5`
