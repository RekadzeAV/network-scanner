# Changelog

Все значимые изменения в проекте Network Scanner будут документироваться в этом файле.

Формат основан на [Keep a Changelog](https://keepachangelog.com/ru/1.0.0/),
и проект придерживается [Semantic Versioning](https://semver.org/lang/ru/).

## [Unreleased]

### Инженерный baseline
- Базовые инженерные точки входа в `Makefile`: `make build`, `make test`, `make run`, `make deploy`.
- Скрипты первичной настройки окружения: `scripts/bootstrap.sh` и `scripts/bootstrap.ps1`.
- Локальная интеграционная проверка: `scripts/integration-check.sh` и `scripts/integration-check.ps1`.
- Канонический роадмап `docs/ROADMAP.md`.
- Документация развёртывания и отката: `docs/deployment.md#rollback`.
- ADR-решение по инженерному baseline: `docs/adr/0001-project-governance.md`.
- Руководство по вкладу: `CONTRIBUTING.md`.
- Пример переменных окружения: `config/config.example.env`.
- Базовая pre-commit конфигурация: `.pre-commit-config.yaml`.

### Добавлено (функциональность и платформа)
- Формализован closure P1 для macOS/Linux через обязательный CI job `P1 Closure (Unix)` (матрица `ubuntu-latest` + `macos-latest`) с запуском `./scripts/p1-closure-check.sh`.

- **Этап 1 / P3 (кроссплатформенность, тесты, UX/perf):**
  - Единый execution-слой для внешних утилит с нормализованными кодами ошибок (`not_installed`, `permission_denied`, `timeout`, `network_error`, `unknown`)
  - Унификация пользовательских сообщений ошибок tool-режимов в CLI/GUI
  - CI workflow: lint -> test (Win/Linux/macOS) -> build+smoke + загрузка артефактов
  - Конфигурация линтера в репозитории: `.golangci.yml`
  - Integration-тесты под build tag `integration` + Makefile цель `test-integration`
  - Golden-тесты форматированного вывода (`internal/display/testdata/*.golden.txt`) с обновлением через `UPDATE_GOLDEN=1`
  - Скрипты формального закрытия P3: `scripts/p3-closure-check.sh` и `scripts/p3-closure-check.ps1`
  - Makefile цели для P3: `p3-check`, `p3-check-win`
  - Предупреждение/подтверждение запуска для крупных подсетей в GUI
  - Валидация и ограничение `threads` в диапазоне `1..512` для CLI/GUI
  - Троттлинг UI-обновлений прогресса сканирования в GUI
  - Персистентность рекомендованного GUI-профиля усилена: класс бейджа (`small/medium/large/very-large`) сохраняется отдельно и восстанавливается при старте (с fallback на legacy текст)
  - Бенчмарк `BenchmarkFormatResultsAsTextLarge` и baseline документ `docs/P3_PERF_BASELINE.md`

- **Расширенное управление отображением результатов в GUI (`Сканирование`):**
  - Сортировка по `IP` / `HostName`
  - Лимит отображаемых чипов портов (`12/24/48`)
  - Быстрые фильтры по типам устройств и опция `Только с открытыми портами`
  - Индикатор количества активных фильтров и кнопки быстрого сброса
  - Подрежимы `Devices` / `Security` в UI результатов
  - `Host Details Drawer` с быстрыми действиями (`Ping`, `Traceroute`, `DNS`, `Whois`, `Wake-on-LAN`)
  - Явная встроенная аналитика в render pipeline (`markdown summary` для таблицы, `pie charts` для карточек)
  - `Security Dashboard` (агрегация `audit + risk signatures`, таблица findings, экспорт HTML)
  - Адаптивные профили layout (`compact/normal/wide`) по ширине окна для устойчивого поведения в windowed/fullscreen
  - Runtime-перестройка grid-блоков фильтров/сортировки/operations в compact режиме (1-2 колонки)
  - Адаптивный рендер `Host Details` (вертикальный split в compact) и укороченные заголовки/колонки таблицы
  - Панели управления вкладок `Сканирование`, `Топология`, `Инструменты` переведены на `VScroll` для малых высот окна
- **Operations Runtime в GUI инструментах:**
  - Добавлен `OperationsManager` (`queued/running/success/failed/canceled`) с `Run/Cancel/Retry/List/Subscribe`
  - Во вкладке `Инструменты` добавлен `Operations Center` (история операций + `Retry/Cancel`)
  - Запуски инструментов переведены на единый runtime операций
- **Декомпозиция GUI orchestrator-логики:**
  - Добавлены `internal/gui/scan_controller.go` и `internal/gui/topology_controller.go` для вынесения UI-state переходов scan/topology
- **Персистентность настроек отображения результатов:**
  - Сохраняются и восстанавливаются режим, сортировка, лимит чипов, строка фильтра и быстрые фильтры
  - Добавлено сохранение/восстановление подрежима результатов (`Devices`/`Security`)
- **Экспорт текущего представления из GUI:**
  - `Сохранить результаты` теперь экспортирует отфильтрованный и отсортированный набор устройств (то, что пользователь видит в UI)

- **Этап 1 / P2 (WOL, баннеры, определение ОС):**
  - Wake-on-LAN в CLI: `--wol-mac`, `--wol-broadcast`
  - Выбор интерфейса для WOL в CLI: `--wol-iface` (автоопределение broadcast)
  - Wake-on-LAN в GUI вкладке `Инструменты`
  - Сбор баннеров/версий сервисов: `--grab-banners`
  - Отдельный `Version` в модели порта + вывод в GUI/CLI/JSON
  - Управление показом сырого баннера в CLI: `--show-raw-banners`
  - Переключатель показа raw banner в GUI (с сохранением настройки)
  - Active-режим определения ОС: `--os-detect-active` + переключатель в GUI
  - Добавлено поле `GuessOSReason` (обоснование эвристики) в вывод и экспорт
  - Расширены active-сигнатуры ОС (Windows Server/WinRM, Linux Docker/K8s, Apple mDNS, Android debug)
  - Добавлены unit-тесты для `internal/wol` и `internal/banner` (резолв WOL/broadcast, парсинг версии из баннеров)
  - Успешно пройден локальный closure-прогон на Windows: `.\scripts\p2-closure-check.ps1` (включает `go test ./...`, smoke и sanity P2-флагов)
  - Усилены PowerShell smoke-скрипты: строгая проверка exit code нативных команд и сборка smoke-бинарника во временный `.exe` (устранение flaky-блокировок файла)
  - Усилены Unix smoke-скрипты: временный smoke-бинарник + безопасный cleanup через `trap` (снижение риска конфликтов файла)
  - Обновлены release/closure чеклисты: добавлены явные команды кросс-ОС прогона для Linux/macOS
  - В checklist зафиксирован операционный статус кросс-ОС прогона: Windows подтвержден, Linux/macOS ожидают запуск в целевых средах
  - Добавлены `p2-closure-check` скрипты (Unix/Windows) и Makefile-цели (`p2-check`, `p2-check-win`) для воспроизводимой проверки P2-флагов (баннеры/OS-active/WOL validation)
- **Этап 2 / P2 (управление оборудованием + сигнатуры рисков):**
  - Добавлен модуль `internal/risksignature` с versioned локальной базой сигнатур домашних рисков и explain-выводом причин срабатывания
  - Добавлен CLI флаг `--risk-signatures` для анализа результатов сканирования по сигнатурам рисков
  - Добавлен модуль `internal/devicecontrol` (MVP) для действий `status/reboot` по HTTP API с вендор-адаптерами
  - Добавлены CLI флаги device-control: `--device-action`, `--device-target`, `--device-vendor`, `--device-user`, `--device-pass`, `--device-confirm`, `--device-timeout`, `--audit-log`
  - Для reboot добавлено явное подтверждение `--device-confirm I_UNDERSTAND`
  - Добавлен JSONL audit trail для device-control действий (`audit/device-actions.log`)
  - GUI вкладка `Инструменты` расширена кнопками `Risk Signatures`, `Device Status`, `Device Reboot` и полями `Device Control`
  - В GUI добавлено подтверждение опасного reboot-действия и сохранение device-control параметров в `Preferences`
  - Добавлены вендор-профили device-control: `generic-http`, `tp-link-http`
  - Security HTML report расширен секцией `Risk Signature Findings` (вместе с `CVE Findings`)
  - Добавлены API `RenderSecurityHTMLWithRisk` / `SaveSecurityHTMLWithRisk`
  - Усилен audit-контур security report: metadata-блок (`report-id`, `mode`, `policy`, `unsafe-consent`), явный индикатор `REDACTION: ON|OFF`, warning для unredacted-режима
  - Добавлено автоименование report-файла с `report-id` (`--security-report-file auto`) и вывод `report-id` в CLI после сохранения
  - CI workflow расширен job `Stage2 P3 Closure` (матрица `ubuntu-latest` + `windows-latest`)
  - Скрипты CI статуса/триггера (`check-ci-status.*`, `trigger-ci-workflow.*`) переведены на strict-проверку `Stage2 P1 Closure` + `Stage2 P3 Closure` с fail-fast при незелёных jobs
  - Добавлены unit-тесты для risk signatures, device-control и security report с risk-section
  - Добавлены closure-скрипты Stage2/P2 (`scripts/stage2-p2-closure-check.sh` / `.ps1`) и Makefile-цели (`stage2-p2-check`, `stage2-p2-check-win`)
- **Этап 1 / P1 (инструменты и фильтры):**
- **Инструменты P1 в CLI (`cmd/network-scanner`):**
  - Режимы `--ping`, `--traceroute`, `--dns` (без запуска полного сканирования)
  - Параметры `--dns-server`, `--raw`, `--ping-count`, `--tool-timeout`, `--traceroute-max-hops`
  - Структурная сводка по ping/traceroute + опциональный raw-вывод
- **Этап 2 / P1 (старт реализации):**
  - Добавлены tool-режимы CLI: `--whois`, `--wifi`
  - Добавлен CLI-флаг `--audit-open-ports` для базового аудита рисков по результатам сканирования
  - Добавлен CLI-флаг `--audit-min-severity` для порога критичности (`all|low|medium|high|critical`)
  - В GUI (`Инструменты`) добавлены кнопки `Whois`, `Wi-Fi`, `Аудит портов`
  - В GUI добавлен селектор `Audit min severity` с персистентностью значения
  - Добавлен модуль `internal/audit` с базовыми правилами рисков (telnet/ftp/smb/rdp и др.)
  - Добавлены closure-скрипты Stage2/P1 (`scripts/stage2-p1-closure-check.sh` / `.ps1`) и Makefile-цели (`stage2-p1-check`, `stage2-p1-check-win`)
  - CI workflow расширен job `Stage2 P1 Closure` (матрица `ubuntu-latest` + `windows-latest`)
  - Для `whois` добавлен конфигурируемый RDAP endpoint через `NETWORK_SCANNER_RDAP_BASE_URL` (для детерминированных/offline тестов)
  - Добавлены детерминированные RDAP unit-тесты (`internal/nettools/whois_test.go`) без внешней сети
  - Добавлен CLI e2e тест `TestRunToolsMode_WhoisUsesRDAPFallback` в `cmd/network-scanner/main_test.go`
  - `stage2-p1-closure-check` усилен шагом `go test ./cmd/network-scanner -run Whois` для явной валидации `--whois` в `runToolsMode`
  - Усилен парсинг Wi-Fi Linux (`nmcli`): корректная обработка escaped-разделителей (`\:`) в SSID
  - Усилен парсинг Wi-Fi Windows (`netsh`) для локализованного RU-вывода и нормализации состояния (`connected/disconnected/unknown`)
  - Добавлены unit-тесты Wi-Fi edge-кейсов: SSID с `:`, RU-ключи `netsh`, сценарии `disconnected`/`unknown` и проверка итоговой summary-выдачи
- **Инструменты P1 в GUI (`internal/gui/app.go`):**
  - Вкладка `Инструменты` с кнопками `Ping`, `Traceroute`, `DNS`
  - Поля управления: host, ping count, timeout, traceroute max hops, DNS resolver
  - Сохранение/восстановление настроек инструментов через `fyne.Preferences`
- **Расширенные фильтры P1 в GUI:**
  - Фильтр по CIDR и фильтр по состоянию портов (`all/open/closed/filtered`)
- **Планирование дополнительного развития (D-трек):**
  - В `docs/ROADMAP_P1_P3.md` добавлен структурированный блок `D1..D4` (Topology/Export/GUI hardening)
  - В `docs/DETAILED_BACKLOG_P3_STAGE2.md` добавлен детальный backlog задач дополнительной реализации
  - Пресеты фильтров (слоты `1/2/3`) с сохранением и применением
- **Регрессионные проверки и тесты P1:**
  - CLI smoke для инструментов: `scripts/smoke-cli-tools.sh` / `.ps1`
  - Unit-тесты парсеров ping/traceroute (`internal/nettools`)
  - Unit-тесты фильтров/пресетов/сценариев сохранения (`internal/gui`)
  - `smoke-cli-tools` переведен на offline-stable DNS проверку (без обязательной зависимости от внешнего `1.1.1.1`)
  - В `smoke-cli-tools` добавлен whois e2e шаг через `go test ./cmd/network-scanner -run WhoisUsesRDAPFallback`

### Изменено
- **Рефакторинг GUI-слоя результатов:**
  - Логика вынесена из `app.go` в `internal/gui/results_view.go` и `internal/gui/results_charts.go`
  - Добавлен кэш ресурсов круговых диаграмм для ускорения повторной отрисовки
- **Стабилизация и UX для heavy-сканов в GUI:**
  - Вывод `Ping/Traceroute` в `Инструментах` переведен на совместимый markdown-формат (без HTML `<details>`), что исправляет некорректное отображение raw-текста в `RichText`
  - Добавлен RU-парсинг статистики Windows `ping` (локализованные строки `потеряно`, `минимальное/максимальное/среднее`)
  - Проверка доступности хостов в `DefaultNetworkProber` переведена на параллельный probe с ранним выходом и короткими probe-timeout
  - В `scanner` добавлен адаптивный лимит параллельных порт-проверок на хост (ограничение суммарной нагрузки)
  - В GUI добавлен управляемый автопрофиль сканирования (чекбокс + подсказка + сохранение в preferences)
  - Добавлен цветовой индикатор состояния автопрофиля в UI (`Автопрофиль: ВКЛ/ВЫКЛ`) и кнопка с пояснением логики автокоррекции

### Документация
- Обновлены `README.md`, `docs/USER_GUIDE.md`, `docs/TECHNICAL.md`, `docs/GUI_SMOKE_CHECKLIST.md` под новые CLI/GUI инструменты, фильтры и поведение сохранения.
- Обновлены `README.md` и `docs/USER_GUIDE.md` по Stage2 P2 (`Risk Signatures`, `Device Control`, troubleshoot и примеры).
- Обновлен `docs/RELEASE_ACCEPTANCE_CHECKLIST.md` проверками Stage2 P2 (GUI flow, confirm reboot, audit log, security report sections).
- Добавлены smoke-скрипты проверки GUI матрицы разрешений: `scripts/smoke-gui-resolution.ps1` и `scripts/smoke-gui-resolution.sh`.
- Добавлен `docs/RELEASE_OPERATIONS_CHEATSHEET.md` с единым набором closure-команд (`P1/P2/P3/Stage2-P1/Stage2-P2`) для релизного дежурства.
- В `docs/RELEASE_OPERATIONS_CHEATSHEET.md` добавлен `6-command runbook (copy/paste)` для быстрого релизного прогона (включая `stage2-p3-closure-check`).
- В `docs/USER_GUIDE.md` добавлен `6-command runbook (copy/paste)` (Unix/PowerShell) и явное включение `p3-closure-check`/`stage2-p3-closure-check` в блок быстрой проверки этапов.
- В `docs/TECHNICAL.md` синхронизирован `6-command runbook (copy/paste)` (Unix/PowerShell) и добавлен `p3-closure-check` в секцию closure-проверок.
- В `README.md` добавлен синхронизированный `6-command runbook` (Unix/PowerShell) с ссылкой на `docs/RELEASE_OPERATIONS_CHEATSHEET.md`.
- В `README.md` добавлена пометка о синхронизации терминов runbook с `docs/USER_GUIDE.md` и `docs/TECHNICAL.md`.
- Синхронизированы `docs/README.md`, `docs/TECHNICAL.md`, `docs/USER_GUIDE.md` под усиленные Stage2/P1 проверки (`whois` в smoke/closure и RDAP fallback e2e).
- Добавлен краткий релиз-итог `docs/RELEASE_SUMMARY_UI_RESULTS.md`.
- Добавлена дорожная карта приоритетов P1–P3: `docs/ROADMAP_P1_P3.md` (ссылка в корневом `README.md`).
- Добавлен детализированный backlog по этапам: `docs/DETAILED_BACKLOG_P3_STAGE2.md`.
- Для D-track hardening добавлены и синхронизированы документы:
  - `docs/GRAPHML_COMPATIBILITY_CHECK.md` — протокол ручной проверки импорта GraphML (yEd/Gephi)
  - `docs/D_TRACK_EVIDENCE_TEMPLATE.md` — стандартный evidence-блок для PR/release
  - обновлены ссылки и инструкции в `README.md`, `docs/USER_GUIDE.md`, `docs/TECHNICAL.md`, `docs/RELEASE_ACCEPTANCE_CHECKLIST.md`.
- Обновлен `docs/RELEASE_OPERATIONS_CHEATSHEET.md`: добавлен раздел `Stage2/P3 Sign-off Commands` (closure -> ci-status -> trigger-ci -> finalize-signoff) для Unix/Windows.
- Добавлен `docs/RELEASE_SUMMARY_STAGE2_P3.md` с кратким статусом закрытия Stage2/P3 и списком remaining шагов до formal close.
- Добавлен `docs/RELEASE_READINESS_STAGE2_P3_PR_SNIPPET.md` (ready-to-paste PR block для Stage2/P3, EN/RU short/long).
- Добавлен `docs/FINAL_PR_COMMENT_STAGE2_P3_READY.md` (готовый финальный PR-комментарий по Stage2/P3, RU).

### D-Track evidence (template for release notes)

При релизе с изменениями topology/export hardening рекомендуется фиксировать:

- Smoke/CI:
  - PASS `smoke-d-track-topology-export` (`.sh/.ps1`)
  - PASS job `D-Track Smoke (Topology Export)`
  - URL CI run
- Export consistency:
  - эквивалентность множеств узлов/связей между `json` и `graphml`
  - наличие `source_type`/`confidence`/`evidence` в GraphML
- Graphviz fallback:
  - корректный экспорт `png/svg` при наличии `dot`
  - fallback в `json` при отсутствии `dot`
- External compatibility:
  - yEd import PASS
  - Gephi import PASS

### D-Track Evidence (current snapshot)

- Smoke / CI:
  - [x] Windows smoke: `.\scripts\smoke-d-track-topology-export.ps1` PASS
  - [ ] Linux/macOS smoke: `./scripts/smoke-d-track-topology-export.sh` (pending)
  - [ ] CI job `D-Track Smoke (Topology Export)` (pending)
  - [ ] CI URL: pending
- Export Consistency (`json` vs `graphml`):
  - [x] Node-set equivalence: PASS
  - [x] Edge-set equivalence: PASS
  - [x] GraphML metadata keys present: `source_type`, `confidence`, `evidence`
- Graphviz Fallback:
  - [x] Without `dot`: command does not fail, fallback JSON is generated, diagnostic message is present
  - [ ] With `dot`: direct `png/svg` generation (pending)
- External Compatibility:
  - [ ] yEd import PASS (pending)
  - [ ] Gephi import PASS (pending)

## [1.0.4] - 2026-03-27

### Добавлено
- **Формализованный SNMP отчет в topology режиме:**
  - Новый API `CollectWithReport(...)` в `internal/snmpcollector`
  - Метрики: `TotalSNMPTargets`, `Connected`, `Partial`, `Failed`
  - Детализация ошибок по устройствам (`connect_error`, `query_error`)
- **Модель происхождения/достоверности связей в топологии:**
  - Поля `Link.SourceType` (`lldp|fdb|inferred`)
  - Поля `Link.Confidence` (`high|medium|low`)
  - Поле `Link.Evidence` для диагностируемости
- **Regression smoke-проверки CLI:**
  - `scripts/smoke-cli-no-topology.sh` / `.ps1`
  - `scripts/smoke-cli-topology.sh` / `.ps1`

### Изменено
- **Улучшена дедупликация связей LLDP/FDB:**
  - При конфликте для одной пары endpoint выбирается связь с более высоким confidence
  - LLDP-связи имеют приоритет над FDB при равном endpoint наборе
- **CLI и GUI переведены на расширенный SNMP flow:**
  - Используется `CollectWithReport(...)` вместо "немого" сбора
  - В интерфейсах выводится итог SNMP-опроса (ok/partial/failed)
- **Документация синхронизирована с текущей реализацией:**
  - Обновлены `README.md`, `docs/USER_GUIDE.md`, `docs/TECHNICAL.md`, `docs/ARCHITECTURE.md`

### Исправлено
- **Фильтрация MAC-адресов в FDB-логике топологии:**
  - Исправлена обработка multicast через проверку I/G бита первого байта
  - Отфильтрованы `broadcast` и `00:00:00:00:00:00`
  - Исключены self-MAC связи коммутатора
- **Устойчивость GUI при длительном построении топологии:**
  - Блокировка действий построения/сохранения/превью на время SNMP+Build
  - Возврат кнопок в рабочее состояние после завершения/ошибки

## [1.0.3] - 2026-02-27

### Добавлено
- **Улучшенное отображение результатов в GUI:**
  - Результаты сканирования теперь отображаются в табличном формате
  - Таблица устройств с колонками: HostName / IP / MAC / Порты / Протокол / Тип устройства / Производитель
  - Сетевая аналитика в виде таблицы протоколов (Telnet, SSH, FTP, HTTP, HTTPS, SMB и т.д.)
  - Таблица типов устройств (Network Device / Computer / Server)
- **Новые режимы работы в консольном приложении:**
  - Режим 1: Не закрывать окно по окончании процесса сканирования (ожидание нажатия Enter)
  - Режим 2: Автоматическое сохранение результатов рядом с исполняемым файлом в формате txt
  - Интерактивный выбор режима при запуске приложения

### Исправлено
- **Критические исправления в скриптах сборки:**
  - Добавлен `CGO_ENABLED=1` для всех GUI сборок (macOS, Linux) - Fyne требует CGO
  - Исправлен скрипт `build-macos.sh`: добавлен CGO для GUI и создание universal binary для GUI
  - Исправлен скрипт `build.sh`: добавлен CGO для всех GUI сборок
  - Исправлен `go.mod`: `github.com/jung-kurt/gofpdf/v2` теперь прямая зависимость
- **Обновлена документация:**
  - Обновлена версия в "Инструкция по эксплуатации.md" до 1.0.3
  - Исправлены несоответствия версий в документации

### Улучшено
- Расширен маппинг типов устройств для корректной группировки в аналитике
- Улучшена структура отображения результатов (таблицы вместо списков)
- Имена файлов результатов включают дату и время сканирования
- Улучшена обработка ошибок в скриптах сборки для macOS

## [1.0.2] - 2026-02-27

### Исправлено
- **Критическое исправление зависаний в Windows:**
  - Исправлено зависание при сканировании портов (этап 2)
  - Добавлены таймауты для всех сетевых операций:
    - DNS запросы (`net.LookupAddr`) - таймаут 2 секунды
    - ARP команды в Windows (`arp -a`) - таймаут 3 секунды
    - Получение сетевых интерфейсов (`net.Interfaces`) - таймаут 3-5 секунд
    - Получение адресов интерфейсов (`iface.Addrs`) - таймаут 1-2 секунды
  - MAC адрес и hostname теперь получаются асинхронно и не блокируют сканирование портов
  - Улучшена функция `IsPortOpen` с использованием `net.Dialer` для лучшей работы в Windows
  - Добавлена более частая проверка контекста отмены в цикле сканирования портов
  - Ограничена частота обновления прогресса (каждые 5 хостов) для предотвращения блокировки UI
  - Улучшено закрытие TCP соединений

## [1.0.1] - 2026-02-27

### Исправлено
- Исправлена ошибка "Error in Fyne call thread" при проверке доступности хостов
  - Все обновления UI из горутин теперь обернуты в `fyne.Do()` для выполнения в главном потоке
  - Исправлены обновления прогресса сканирования
  - Исправлены обновления результатов сканирования
  - Исправлены периодические обновления UI
  - Исправлена обработка таймаутов
  - Исправлено автоопределение сети

## [1.0.0] - 2024-12-27

### Добавлено
- CLI версия сканера локальной сети
- GUI версия с графическим интерфейсом (Fyne)
- Автоматическое определение локальной сети
- Сканирование активных хостов
- Сканирование TCP портов
- Определение MAC адресов через ARP
- Определение типов устройств по открытым портам
- Определение производителя по MAC адресу (OUI)
- Детальная аналитика по протоколам и портам
- Красивый табличный вывод результатов
- Сохранение результатов в файл (GUI версия)
- Кроссплатформенная поддержка (Windows, macOS, Linux)
- Многопоточное сканирование с настраиваемым количеством потоков
- Обработка сигналов для корректного завершения
- Подробная документация

### Технические детали
- Использование Go 1.21+ (на момент релиза 1.0.0)
- Библиотеки: gopacket, go-pretty, fyne
- Модульная архитектура
- Параллельное сканирование через горутины

### Известные ограничения
- Только TCP порты (UDP не поддерживается на момент релиза 1.0.0)
- MAC адреса могут не определяться без прав администратора
- Эвристическое определение типов устройств

