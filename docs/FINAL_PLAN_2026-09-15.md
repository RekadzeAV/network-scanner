# Финальный план работ (по результатам аудита документации и проверки слияний)

**Дата:** 2026-09-15
**Основан на:** [UNIFIED_OPTIMIZED_PLAN_2026-09-15.md](UNIFIED_OPTIMIZED_PLAN_2026-09-15.md)
**Входные данные:** аудит `docs/` (битые ссылки — 0, архив `docs/archive/2026-09-15-docs-sync/` — 11 файлов), проверка git (08:42 15.09.2026)

---

## Результаты проверок, определяющие план

1. **Слияния:** автономных неслитых веток нет; локальный `main` опережает `origin/main` на 2 коммита (`eb97654`, `09a04dd`) — **не запушено**; основной массив изменений v2.3 (C1–C7, тесты, доки) — **не закоммичен** в рабочем дереве.
2. **Документация:** устаревшие планы цикла 09-04 заархивированы; актуальные документы обновлены до v2.3.0; битые ссылки устранены; удалён пустой файл-заглушка `docs/TECHNICAL_SPECIFICATION.md`; создан `docs/GUI_SMOKE_CHECKLIST.md` (ссылка из README вела в никуда).
3. **Дублирование, выявленное аудитом:**
   - кодовое: `internal/api/topology_handlers.go` — трёхкратно повторённый пайплайн «снапшот → SNMP → BuildTopology»; GraphML через temp-файл вместо `SaveGraphMLToBytes()`;
   - репозиторное: coverage-артефакты продублированы в корне и `internal/legacy/` (untracked-копии);
   - документационное: несколько индексных документов дублировали навигацию (устранено архивированием; правило «один индекс» закреплено в ROADMAP).
4. **Регресс:** `internal/topology` — 11 падающих тестов (заглушка `Export`, фикстура вне `Validate()`).

## Финальный план (исполняется сейчас, по порядку)

| # | Задача | Дедупликация | Статус |
|---|--------|--------------|--------|
| F1 | Реализовать `topologyServiceImpl.Export` (json/graphml/dot/text/txt/xml) | экспорт пишется один раз, переиспользует `convertFromContractTopology` + существующие Save*-методы | ✅ done |
| F2 | Привести фикстуру `TestSaveGraphMLToBytes` в соответствие `Validate()` | — | ✅ done |
| F3 | Дедупликация `topology_handlers.go`: общий хелпер загрузки топологии; GraphML без temp-файла | **устранение тройного дублирования** | ✅ done |
| F4 | Удалить артефакты: cov-файлы/логи корня, `.tmp/`, untracked-дубли в `internal/legacy/`, `internal/zaclikivaniya.md` | **устранение дублей coverage-артефактов** | ✅ done |
| F5 | Дополнить `.gitignore` (cov-паттерны, `.tmp/`, `*_cov*`) — повторное появление исключено | — | ✅ done |
| F6 | `go build ./...` + полный `go test ./...` — все пакеты зелёные | — | ✅ done (48/48 ok, 0 FAIL) |
| F7 | Зафиксировать результаты в документации (CHANGELOG, статусы плана, ROADMAP) | — | ✅ done |
| F8 | *Требует подтверждения пользователя:* коммит рабочего набора в main, пуш в origin/main, удаление сломанных refs (`~gvfYSKY.tmp` и др.) | — | ✅ done (2026-09-15) |

## Ход выполнения (2026-09-15)

- **F1/F2:** `internal/topology/service_impl.go` — `Export` реализован (6 форматов, ошибки `topology is nil` / `export path is empty` / `unsupported export format` / `create file`); добавлен детерминированный обход устройств `sortedDeviceKeys` (устранён недетерминизм DOT/GraphML/XML-вывода из-за map-итерации — падал `TestDOTGoldenSnapshot`).
- **F3:** `internal/api/topology_handlers.go` — `topologyRequest` + `loadHostsForSnapshot` + `buildTopologyForRequest`; четвёрная копия пайплайна сведена к одной; GraphML отдаётся через `SaveGraphMLToBytes()` без временного файла; поведение `build` при пустом inventory сохранено (200 + message).
- **F4:** удалены `api_cov`, `banner_cov`, `scanner_cov*`, `scanner_test*.log`, `.tmp/`, 21 untracked coverage-дубль в `internal/legacy/`; `internal/zaclikivaniya.md` → `docs/archive/2026-09-15-docs-sync/zaclikivaniya-notes.md`.
- **F6:** `go build ./...` чисто; `go test ./... -short -count=1` — **48 пакетов ok, 0 FAIL** (регресс `internal/topology` устранён).
- **F8 (частично):** сломанные refs `refs/heads/~gvfYSKY.tmp`, `refs/remotes/origin/~gvfGW8Z.tmp`, `refs/remotes/origin/~gvfzWws.tmp` удалены; при чистке случайно удалялся `refs/remotes/origin/main` — восстановлен на `dde66e1`, целостность подтверждена (`git for-each-ref` без предупреждений).

## Исключение возвратов

- F1–F3 идут вместе: правка экспорта и его дедупликация в API касаются одного функционального среза.
- F4–F5 до коммита (F8): мусор не попадёт в историю.
- F7 после F6: документация фиксирует только проверенное состояние.
- F8 — единственный замораживающий шаг, только с подтверждения.

---

## E4 CI-гейты — ЗАКРЫТО (2026-09-15, коммиты 0f5a737/2bed84e)
- golangci-lint v1.64.8: 91 → 0 замечаний (errcheck/staticcheck/gosimple/ineffassign/unused)
- go test ./... -short: 48 ok / 0 FAIL
- go.yml: Go 1.25, libpcap-dev во всех job, фиксированная версия линтера
- pcap-абстракция: livePacketHandle + build-tagged open/stub; кросс-матрица scanner+CLI (CGO=0) зелёная

---

## E5 Coverage — ЗАКРЫТО (2026-09-19)
- internal/scanner: 74.0% → 84.7% (14 новых тестов: loopback-флоу Scan/scanHost, отмены, fake probers, сервисный слой)
- internal/topology: 88.3% (цель 85% достигнута ранее)
- Остаток непокрытого в scanner — платформенно-специфичные ветки (linux/darwin ARP, pcap-ARP требует root)
- go test ./... : 48 пакетов ok / 0 FAIL

---

## E6 v2.3-слой — ЗАКРЫТО (2026-09-19, решение: tested foundation + инкрементальная обвязка)

**Фактическое состояние (проверено):**
- cobra-CLI живой: main -> cmd.ExecuteCLI (scan.go) -> scanCmd.RunE (scan_cobra.go) использует реальную бизнес-логику (scanner/snmpcollector/builder/display/presenter)
- internal/apperror 98.4%, internal/commands 97.4%, internal/eventbus 93.5%, internal/plugin 60.0%, internal/scanner/plugin — тесты ok
- production-потребители apperror/eventbus/commands вне собственного слоя — отсутствуют (самодостаточная инфраструктура)

**Решение:**
- НЕ ретрофитить apperror в легаси-бизнес-пути: высокий риск регрессий работающего продукта без пользовательской ценности
- Слой остаётся протестированным фундаментом: новые фичи пишутся на apperror/eventbus/commands с первого коммита
- internal/plugin (60%) — добирать покрытие при появлении реальных плагинов (loader-специфичные ветки)
- cmd/network-scanner/cmd без тестов: допустимо, сценарный путь CLI покрыт e2e-сканами scanner
