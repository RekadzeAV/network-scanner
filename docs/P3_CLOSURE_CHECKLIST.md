# P3 Closure Checklist (Stage 1)

Цель: формально закрыть `Этап 1 / P3` (кроссплатформенность, тесты, UX/perf) единым воспроизводимым прогоном.

---

## Как заполнять

- `Owner` - ответственный за задачу.
- `ETA` - плановая дата завершения (`YYYY-MM-DD`).
- `Status` - одно из: `todo`, `in_progress`, `blocked`, `done`.
- Отмечайте чекбокс `[x]` только после выполнения критериев приёмки.

---

## 1) Кроссплатформенный execution-слой

### P3-01 Единый executor внешних команд

- [x] Реализовать общий executor с результатом `stdout/stderr/exitCode/duration`.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Добавить поддержку `timeout` и `cancel` без зависаний.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Подключить executor ко всем инструментальным вызовам в CLI/GUI.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`

**Критерий приёмки:** все внешние инструменты запускаются через единый execution-слой.

### P3-02 OS-адаптеры аргументов

- [x] Разделить аргументы `ping` для Windows и Unix-подобных ОС.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Разделить аргументы `traceroute/tracert` по ОС.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Проверить корректность max-hops/timeout/count на Win/macOS/Linux.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`

**Критерий приёмки:** нет сбоев из-за неподдерживаемых флагов на целевых ОС.

### P3-03 Preflight-проверка утилит

- [x] Добавить проверку наличия внешней команды в `PATH` до запуска.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Добавить типизированную ошибку `not_installed`.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Показать понятный fallback в GUI/CLI без падения.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`

**Критерий приёмки:** при отсутствии утилит пользователь получает понятную диагностику.

### P3-04 Нормализация ошибок и статусов

- [x] Ввести единый набор ошибок: `not_installed`, `permission_denied`, `timeout`, `network_error`, `parse_error`, `unknown`.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Добавить маппинг системных ошибок в единые коды.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Унифицировать пользовательские сообщения для CLI/GUI.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`

**Критерий приёмки:** одинаковые типы ошибок отображаются одинаково по смыслу в обоих интерфейсах.

---

## 2) CI и инженерные проверки

### P3-05 Matrix CI на 3 ОС

- [x] Настроить matrix job для `windows`, `ubuntu`, `macos`.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Добавить шаги `go test ./...` и `go build ./...`.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [ ] Проверить стабильность прогонов на всех ОС.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `in_progress`

**Критерий приёмки:** pipeline стабильно зелёный на 3 ОС.

### P3-06 Линтер и quality gate

- [x] Подключить линтер в CI как обязательную проверку PR.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Зафиксировать конфигурацию линтера в репозитории.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`

**Критерий приёмки:** PR не может быть смержен при провале lint-проверки.

### P3-07 Smoke-build GUI/CLI артефактов

- [x] Добавить сборку CLI и GUI бинарей в CI.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Добавить минимальную smoke-проверку запуска бинарей.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Публиковать артефакты сборки по ОС.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`

**Критерий приёмки:** доступные артефакты CLI/GUI для каждой целевой ОС.

---

## 3) Тестирование

### P3-08 Unit-тесты `internal/nettools`

- [x] Добавить кейсы парсинга stdout/stderr для Win/Unix.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Добавить кейсы timeout/cancel.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Добавить кейсы нормализации ошибок.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`

**Критерий приёмки:** регрессии парсинга и таймаутов ловятся unit-тестами.

### P3-09 Unit-тесты `internal/osdetect`, `internal/banner`, `internal/wol`

- [x] Расширить тесты `internal/osdetect` (confidence/reason и граничные случаи).
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Расширить тесты `internal/banner` (sanitize/timeout/empty response).
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Добавить/расширить тесты `internal/wol` (валидный/невалидный MAC, broadcast).
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`

**Критерий приёмки:** у модулей есть стабильные happy-path и negative-path кейсы.

### P3-10 Integration-тесты под тегом

- [x] Вынести integration-тесты под build tag `integration`.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Добавить инструкции запуска integration для Win/macOS/Linux.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Добавить корректные `skip`-условия для неподготовленного окружения.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`

**Критерий приёмки:** integration-прогоны воспроизводимы и не ломают базовый CI.

### P3-11 Golden-тесты выводов

- [x] Зафиксировать golden-файлы для стабильного CLI-формата вывода.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Добавить механизм контролируемого обновления golden-снимков.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`

**Критерий приёмки:** неожиданные изменения форматированного вывода детектятся автоматически.

---

## 4) UX и производительность

### P3-12 Защита от больших подсетей

- [x] Добавить предупреждение для крупных подсетей (например, `/23` и крупнее).
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Добавить явное подтверждение запуска в GUI для тяжёлых диапазонов.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`

**Критерий приёмки:** пользователь предупреждён о риске долгого/тяжёлого сканирования до запуска.

### P3-13 Валидация параллелизма

- [x] Ввести минимальные и максимальные значения concurrency.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Добавить понятные ошибки валидации в CLI/GUI.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`

**Критерий приёмки:** невалидные настройки не приводят к запуску сканирования.

### P3-14 Стабилизация обновления прогресса в GUI

- [x] Добавить throttling/debounce обновлений прогресса.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Проверить отзывчивость GUI на сценарии `/24`.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`

**Критерий приёмки:** во время скана UI остаётся отзывчивым, без заметных фризов.

### P3-15 Perf baseline и budget

- [x] Замерить baseline для `/24` и `/23` (время скана, память, рендер).
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Зафиксировать целевые пороги (perf budget) в документации.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`

**Критерий приёмки:** есть документированный baseline для сравнения регрессий.

---

## 5) Документация и формальное закрытие P3

### P3-16 Обновление `docs/TECHNICAL.md`

- [x] Добавить таблицу ограничений по ОС (права/firewall/raw sockets/ICMP).
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Добавить диагностику типовых ошибок execution-слоя.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`

**Критерий приёмки:** технические ограничения описаны прозрачно и полно.

### P3-17 Обновление `docs/USER_GUIDE.md`

- [x] Добавить раздел troubleshooting по платформам.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Добавить практические сценарии решения ошибок `not_installed/permission_denied/timeout`.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`

**Критерий приёмки:** пользователь может самостоятельно диагностировать и устранить типовые проблемы.

### P3-18 Финальный closure-check

- [ ] Проверить, что CI (`test/build/lint`) зелёный на трёх ОС.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `in_progress`
- [x] Подтвердить прохождение unit/integration/golden сценариев.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Добавить воспроизводимый локальный closure-прогон (`p3-closure-check`) для Unix/Windows.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Подтвердить UX/perf baseline и обновление документации.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`
- [x] Обновить дорожную карту и changelog по результатам закрытия P3.
  - Owner: `TBD`
  - ETA: `TBD`
  - Status: `done`

**Критерий приёмки:** все разделы чеклиста закрыты, `Этап 1 / P3` формально завершён.

---

## P3 Final Sign-off

Заполняется после завершения последнего пункта `P3-18`.

- Статус P3: `in_progress (awaiting CI evidence)`
- Дата финального подтверждения: `TBD`
- Подтверждено (Owner): `TBD`
- CI run URL (green): `TBD`
- Ссылка на release checklist / smoke evidence: `TBD`
- Примечание: для API-скриптов CI требуется `GITHUB_TOKEN`.

Минимальные шаги для финального закрытия:
1. Запустить CI workflow `CI` (UI GitHub Actions или `.\scripts\trigger-ci-workflow.ps1` / `./scripts/trigger-ci-workflow.sh`).
2. Проверить, что required jobs зелёные: `Lint`, `Test*`, `Build and Smoke*`, `Stage2 P1 Closure`, `Stage2 P3 Closure`.
3. Подтвердить green run командой `.\scripts\check-ci-status.ps1` / `./scripts/check-ci-status.sh` (или `make ci-status-win` / `make ci-status`) и внести URL в поле `CI run URL`.
4. Перевести статус в `closed` и заполнить `Дата`/`Owner` (или использовать `.\scripts\finalize-p3-signoff.ps1 -ConfirmedBy "<owner>"` / `./scripts/finalize-p3-signoff.sh ...` / `make p3-signoff-win` / `make p3-signoff`).

Быстрый путь (одной командой):
- Windows: `make p3-close-all-win` (после установки `GITHUB_TOKEN`)
- Linux/macOS: `make p3-close-all` (после установки `GITHUB_TOKEN`)

Критерий перевода статуса в `closed`:
- `P3-18` полностью закрыт;
- в CI подтверждены required jobs на `Windows/macOS/Linux`;
- ручной smoke/UX прогон на целевых ОС зафиксирован.

---

## Критерий "P3 Closed"

Этап считается закрытым, когда:

1. Все пункты `P3-01`...`P3-18` выполнены и отмечены.
2. CI стабильно зелёный на `Windows/Linux/macOS`.
3. Нет критических дефектов по кроссплатформенности, UX и производительности.
4. Ручной smoke/UX прогон выполнен минимум на основной целевой ОС релиза.
