# Release Ready Gap List

Короткий список того, что осталось до полного release sign-off после закрытия фаз Stage 1 и Stage 2 по Windows closure-check.

## Приоритеты закрытия

- `P0` (блокирует sign-off): Cross-OS closure + CI evidence.
- `P1` (обязательная ручная приемка): Manual GUI acceptance.
- `P2` (финальная совместимость): Topology manual checks (включая GraphML import).

## Рекомендуемый порядок выполнения

1. Закрыть `P0` / Cross-OS closure.
2. Закрыть `P0` / CI evidence.
3. Закрыть `P1` / Manual GUI acceptance.
4. Закрыть `P2` / Topology manual checks.

## 1) Cross-OS closure (Linux/macOS) — `P0`

- [ ] Stage 1: выполнить `p1/p2` closure на Linux.
- [ ] Stage 1: выполнить `p1/p2` closure на macOS.
- [ ] Stage 2: выполнить `stage2-p1` closure на Linux/macOS.
- [ ] Stage 2: выполнить `stage2-p2` closure на Linux/macOS.
- [ ] Stage 2: выполнить `stage2-p3` closure на Linux/macOS.

## 2) CI evidence (GitHub Actions) — `P0`

- [ ] Получить успешный run workflow `CI` (required jobs).
- [ ] Подтвердить green matrix на `Windows/Linux/macOS`.
- [ ] Подтвердить green jobs `Stage2 P1 Closure` и `Stage2 P3 Closure`.
- [ ] Внести URL успешного CI run в [P3_CLOSURE_CHECKLIST.md](P3_CLOSURE_CHECKLIST.md).

## 3) Manual GUI acceptance — `P1`

- [ ] Пройти [GUI_SMOKE_CHECKLIST.md](GUI_SMOKE_CHECKLIST.md).
- [ ] Проверить `Host Details Drawer`, `Security Dashboard`, `Operations Center`.
- [ ] Проверить инструменты `Whois/Wi-Fi/Audit`, `Risk Signatures`, `Device Control`.
- [ ] Проверить корректность экспорта текущего представления (`Сохранить результаты`).

## 4) Topology manual checks — `P2`

- [ ] Ручной smoke вкладки `Топология` (open/build/preview/save/zoom).
- [ ] Проверить import `GraphML` в yEd и Gephi.
- [ ] Зафиксировать результаты в [GRAPHML_COMPATIBILITY_CHECK.md](GRAPHML_COMPATIBILITY_CHECK.md).

## Completion rule

Релиз считается полностью готовым, когда все пункты в этом файле отмечены как `[x]` и синхронно отражены в [RELEASE_ACCEPTANCE_CHECKLIST.md](RELEASE_ACCEPTANCE_CHECKLIST.md).

## Текущий blocker (на момент последнего прогона)

- `GITHUB_TOKEN` не задан в окружении, поэтому нельзя запустить автоматический `p3-close-all` и зафиксировать CI evidence через скрипты.
- Последние публичные run workflow `ci.yml` находятся в состоянии `failure`, успешный run для sign-off отсутствует.

## Как разблокировать за 1 проход

1. Экспортировать `GITHUB_TOKEN` (token с доступом к repo/workflow).
   - Preflight перед запуском: `.\scripts\p0-signoff-preflight.ps1` (или `make p0-preflight-win`).
2. Запустить:
   - Windows: `make p3-close-all-win`
   - Linux/macOS: `make p3-close-all`
3. После успешного прогона перенести URL green-run в:
   - [P3_CLOSURE_CHECKLIST.md](P3_CLOSURE_CHECKLIST.md) (`P3 Final Sign-off`)
   - [RELEASE_ACCEPTANCE_CHECKLIST.md](RELEASE_ACCEPTANCE_CHECKLIST.md) (CI evidence блок)

Полный пошаговый runbook: [P0_SIGNOFF_RUNBOOK.md](P0_SIGNOFF_RUNBOOK.md).
Быстрый операционный набор команд и расположение локальных релизных бинарников (`build/release/`): [RELEASE_OPERATIONS_CHEATSHEET.md](RELEASE_OPERATIONS_CHEATSHEET.md); структура каталогов пакета — [BUILD_STRUCTURE.md](BUILD_STRUCTURE.md).
