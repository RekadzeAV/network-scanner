# Checklist Status Index

Единый индекс статусов по ключевым checklist-документам.

## Сводка

| Документ | Статус | Комментарий |
|---|---|---|
| [P1_CLOSURE_CHECKLIST.md](P1_CLOSURE_CHECKLIST.md) | **closed** | Все чекбоксы закрыты. |
| [P3_CLOSURE_CHECKLIST.md](P3_CLOSURE_CHECKLIST.md) | **in_progress** | Остаются CI evidence пункты (cross-OS green). |
| [RELEASE_ACCEPTANCE_CHECKLIST.md](RELEASE_ACCEPTANCE_CHECKLIST.md) | **in_progress** | Остаются cross-OS, CI evidence и ручные GUI/topology проверки. |
| [RELEASE_READY_GAP_LIST.md](RELEASE_READY_GAP_LIST.md) | **in_progress** | Используется как короткий operational backlog до sign-off. |
| [GUI_SMOKE_CHECKLIST.md](GUI_SMOKE_CHECKLIST.md) | **manual** | Пошаговый сценарий ручного прогона (без checkbox-модели). |

## Текущий operational blocker

- `CI evidence` не может быть закрыт автоматически без `GITHUB_TOKEN` в окружении.
- Последний доступный статус `ci.yml`: нет успешного run в недавней истории (только `failure`).

## Что уже закрыто

- Stage 1 (`P1/P2/P3`) доведен до `100%` и синхронизирован в roadmap/snapshot.
- Stage 2 (`P1/P2/P3`) доведен до `100%` и синхронизирован в roadmap/snapshot.
- Windows closure-check подтвержден для:
  - `p1/p2/p3`
  - `stage2-p1/stage2-p2/stage2-p3`

## Что остается до final sign-off

- `P0`: Cross-OS closure на Linux/macOS.
- `P0`: CI evidence (green required jobs + URL успешного run).
- `P1`: Manual GUI smoke по [GUI_SMOKE_CHECKLIST.md](GUI_SMOKE_CHECKLIST.md).
- `P2`: Manual topology checks + GraphML import в yEd/Gephi.

## Источники истины

- Основной финальный чек: [RELEASE_ACCEPTANCE_CHECKLIST.md](RELEASE_ACCEPTANCE_CHECKLIST.md)
- Короткий gap backlog: [RELEASE_READY_GAP_LIST.md](RELEASE_READY_GAP_LIST.md)
- Snapshot текущей готовности: [RELEASE_READINESS_SNAPSHOT.md](RELEASE_READINESS_SNAPSHOT.md)
- Runbook закрытия `P0`: [P0_SIGNOFF_RUNBOOK.md](P0_SIGNOFF_RUNBOOK.md)
- Команды релизного прогона и каталог локальных релизных бинарников (`build/release/`): [RELEASE_OPERATIONS_CHEATSHEET.md](RELEASE_OPERATIONS_CHEATSHEET.md)
- Структура каталогов релизной сборки: [BUILD_STRUCTURE.md](BUILD_STRUCTURE.md)
