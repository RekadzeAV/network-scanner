# P1 Closure Checklist (Stage 1)

Цель: формально закрыть `Этап 1 / P1` (Ping, Traceroute, DNS, фильтры) единым воспроизводимым прогоном.

---

## Quick links

- Release acceptance: [RELEASE_ACCEPTANCE_CHECKLIST.md](RELEASE_ACCEPTANCE_CHECKLIST.md)
- GUI smoke checklist: [GUI_SMOKE_CHECKLIST.md](GUI_SMOKE_CHECKLIST.md)
- Release readiness snapshot: [RELEASE_READINESS_SNAPSHOT.md](RELEASE_READINESS_SNAPSHOT.md)
- Manual sign-off template: [MANUAL_SIGNOFF_TEMPLATE.md](MANUAL_SIGNOFF_TEMPLATE.md)
- Manual sign-off draft: [MANUAL_SIGNOFF_DRAFT.md](MANUAL_SIGNOFF_DRAFT.md)

---

## 1) Базовая проверка сборки и тестов

- [x] `go test ./...` проходит без ошибок.
- [x] Smoke проверки CLI проходят:
  - [x] `./scripts/smoke-cli-no-topology.sh` (или `.ps1`)
  - [x] `./scripts/smoke-cli-topology.sh` (или `.ps1`)
  - [x] `./scripts/smoke-cli-tools.sh` (или `.ps1`)
- [x] Автоматизированный прогон P1:
  - [x] Unix: `./scripts/p1-closure-check.sh`
  - [x] Windows: `.\scripts/p1-closure-check.ps1`

Статус формального closure на Unix-платформах (macOS/Linux):

- [x] В CI добавлен обязательный job `P1 Closure (Unix)` с матрицей `ubuntu-latest` + `macos-latest`.
- [x] Job выполняет `./scripts/p1-closure-check.sh` как единый воспроизводимый критерий закрытия P1.
- [x] P1 на macOS/Linux считается формально закрытым при зелёном статусе job `P1 Closure (Unix)`.

---

## 2) CLI инструменты P1

- [x] `--ping` работает и выводит структурную сводку:
  - [x] `Sent/Received/Loss`
  - [x] `RTT min/avg/max` (если есть данные)
- [x] `--traceroute` работает и выводит hop-сводку.
- [x] `--dns` работает для прямого (`A/AAAA`) и обратного (`PTR`) lookup.
- [x] Параметры инструментов работают:
  - [x] `--ping-count`
  - [x] `--tool-timeout`
  - [x] `--traceroute-max-hops`
  - [x] `--dns-server`
  - [x] `--raw`

Рекомендуемый быстрый набор:

```bash
./network-scanner --ping 127.0.0.1 --ping-count 4 --tool-timeout 15 --raw
./network-scanner --traceroute 8.8.8.8 --traceroute-max-hops 20 --tool-timeout 20
./network-scanner --dns localhost --dns-server 1.1.1.1 --tool-timeout 10 --raw
```

---

## 3) GUI инструменты P1

- [x] На вкладке `Инструменты` доступны кнопки: `Ping`, `Traceroute`, `DNS`.
- [x] Поля параметров работают:
  - [x] Host
  - [x] Ping count
  - [x] Timeout
  - [x] Traceroute hops
  - [x] DNS resolver
- [x] Результаты отображаются без блокировки UI.
- [x] Настройки инструментов сохраняются и восстанавливаются после перезапуска GUI.

---

## 4) Фильтры результатов сканирования (GUI)

- [x] Работает текстовый фильтр.
- [x] Работает фильтр `CIDR`.
- [x] Работает фильтр по состоянию портов (`all/open/closed/filtered`).
- [x] Работают quick-type фильтры и `Только с открытыми портами`.
- [x] Работают пресеты фильтров (слоты `1/2/3`):
  - [x] `Сохранить пресет`
  - [x] `Применить пресет`
- [x] Кнопки `Очистить` и `Сбросить фильтры` ведут себя предсказуемо.

---

## 5) Экспорт с учётом фильтров

- [x] `Сохранить результаты` экспортирует текущий отфильтрованный и отсортированный набор.
- [x] При пустом результате после фильтров показывается понятное сообщение и сохранение не выполняется.
- [x] Регрессионные тесты GUI на save/filter/preset проходят.

---

## 6) Документация и фиксация этапа

- [x] `README.md` содержит актуальные примеры по CLI инструментам P1.
- [x] [USER_GUIDE.md](USER_GUIDE.md) синхронизирован по флагам и GUI вкладке `Инструменты`.
- [x] [TECHNICAL.md](TECHNICAL.md) содержит ограничения инструментов.
- [x] `CHANGELOG.md` отражает все ключевые изменения P1.
- [x] [ROADMAP_P1_P3.md](ROADMAP_P1_P3.md) содержит обновлённый статус P1.

---

## Критерий "P1 Closed"

Этап считается закрытым, когда:

1. Все пункты 1–6 отмечены.
2. Нет критических/блокирующих дефектов в `Ping/Traceroute/DNS/фильтрах`.
3. Ручной UX-прогон выполнен минимум на основной целевой ОС релиза.

---

## Post-P1 продолжение (текущее состояние P2)

Справочный блок для фиксации прогресса после закрытия P1.

### Реализовано

- [x] WOL в CLI (`--wol-mac`, `--wol-broadcast`, `--wol-iface`) и GUI (`Инструменты`).
- [x] Автоподбор broadcast по интерфейсу для WOL (при пустом явном broadcast).
- [x] Сбор баннеров/версий (`--grab-banners`) и отображение в CLI/GUI/JSON.
- [x] Раздельный показ `version` и raw `banner` (`--show-raw-banners` + GUI toggle).
- [x] Active-эвристики определения ОС (`--os-detect-active` + GUI toggle).
- [x] Поля обоснования ОС (`GuessOSReason`) в выводе и экспорте.
- [x] Расширенные active-профили определения ОС + тесты confidence/reason.
- [x] Unit-тесты `internal/osdetect`, `internal/wol`, `internal/banner`.

### Статус формального close P2

- [x] P2 реализован и закрыт в roadmap/status-документах.
- [x] P2 closure-check подтвержден на Windows:
  - [x] `.\scripts\p2-closure-check.ps1` (или `make p2-check-win`)

Рекомендуемые команды для фиксации:

```bash
# Linux/macOS
go test ./...
./scripts/smoke-cli-no-topology.sh
./scripts/smoke-cli-topology.sh
./scripts/smoke-cli-tools.sh
./scripts/p1-closure-check.sh
./scripts/p2-closure-check.sh
```

### Локальная верификация (текущая машина)

- [x] Windows: `go test ./...` проходит.
- [x] Windows: smoke-скрипты (`no-topology`, `topology`, `tools`) проходят.
- [x] Windows: `scripts/p1-closure-check.ps1` проходит.
- [x] Windows: `scripts/p2-closure-check.ps1` проходит.

---

## Stage 2 / P1 (Whois, Wi-Fi, аудит портов) — smoke дополнения

- [x] `scripts/smoke-cli-tools.sh|ps1` включает smoke-проверки:
  - [x] `--audit-open-ports`
  - [x] `--audit-open-ports --audit-min-severity high`
- [x] Добавлен единый closure-скрипт Stage2/P1:
  - [x] Unix: `./scripts/stage2-p1-closure-check.sh` (или `make stage2-p1-check`)
  - [x] Windows: `.\scripts\stage2-p1-closure-check.ps1` (или `make stage2-p1-check-win`)
- [x] `--audit-min-severity` поддерживает значения `all|low|medium|high|critical`.
- [x] В GUI (`Инструменты`) доступен `Audit min severity` с сохранением в Preferences.
