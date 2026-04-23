# P1 Closure Checklist (Stage 1)

Цель: формально закрыть `Этап 1 / P1` (Ping, Traceroute, DNS, фильтры) единым воспроизводимым прогоном.

---

## Quick links

- Release acceptance: `docs/RELEASE_ACCEPTANCE_CHECKLIST.md`
- GUI smoke checklist: `docs/GUI_SMOKE_CHECKLIST.md`
- Release readiness snapshot: `docs/RELEASE_READINESS_SNAPSHOT.md`
- Manual sign-off template: `docs/MANUAL_SIGNOFF_TEMPLATE.md`
- Manual sign-off draft: `docs/MANUAL_SIGNOFF_DRAFT.md`

---

## 1) Базовая проверка сборки и тестов

- [x] `go test ./...` проходит без ошибок.
- [ ] Smoke проверки CLI проходят:
  - [x] `./scripts/smoke-cli-no-topology.sh` (или `.ps1`)
  - [x] `./scripts/smoke-cli-topology.sh` (или `.ps1`)
  - [x] `./scripts/smoke-cli-tools.sh` (или `.ps1`)
- [ ] Автоматизированный прогон P1:
  - [ ] Unix: `./scripts/p1-closure-check.sh`
  - [x] Windows: `.\scripts/p1-closure-check.ps1`

Статус формального closure на Unix-платформах (macOS/Linux):

- [x] В CI добавлен обязательный job `P1 Closure (Unix)` с матрицей `ubuntu-latest` + `macos-latest`.
- [x] Job выполняет `./scripts/p1-closure-check.sh` как единый воспроизводимый критерий закрытия P1.
- [x] P1 на macOS/Linux считается формально закрытым при зелёном статусе job `P1 Closure (Unix)`.

---

## 2) CLI инструменты P1

- [ ] `--ping` работает и выводит структурную сводку:
  - [ ] `Sent/Received/Loss`
  - [ ] `RTT min/avg/max` (если есть данные)
- [ ] `--traceroute` работает и выводит hop-сводку.
- [ ] `--dns` работает для прямого (`A/AAAA`) и обратного (`PTR`) lookup.
- [ ] Параметры инструментов работают:
  - [ ] `--ping-count`
  - [ ] `--tool-timeout`
  - [ ] `--traceroute-max-hops`
  - [ ] `--dns-server`
  - [ ] `--raw`

Рекомендуемый быстрый набор:

```bash
./network-scanner --ping 127.0.0.1 --ping-count 4 --tool-timeout 15 --raw
./network-scanner --traceroute 8.8.8.8 --traceroute-max-hops 20 --tool-timeout 20
./network-scanner --dns localhost --dns-server 1.1.1.1 --tool-timeout 10 --raw
```

---

## 3) GUI инструменты P1

- [ ] На вкладке `Инструменты` доступны кнопки: `Ping`, `Traceroute`, `DNS`.
- [ ] Поля параметров работают:
  - [ ] Host
  - [ ] Ping count
  - [ ] Timeout
  - [ ] Traceroute hops
  - [ ] DNS resolver
- [ ] Результаты отображаются без блокировки UI.
- [ ] Настройки инструментов сохраняются и восстанавливаются после перезапуска GUI.

---

## 4) Фильтры результатов сканирования (GUI)

- [ ] Работает текстовый фильтр.
- [ ] Работает фильтр `CIDR`.
- [ ] Работает фильтр по состоянию портов (`all/open/closed/filtered`).
- [ ] Работают quick-type фильтры и `Только с открытыми портами`.
- [ ] Работают пресеты фильтров (слоты `1/2/3`):
  - [ ] `Сохранить пресет`
  - [ ] `Применить пресет`
- [ ] Кнопки `Очистить` и `Сбросить фильтры` ведут себя предсказуемо.

---

## 5) Экспорт с учётом фильтров

- [ ] `Сохранить результаты` экспортирует текущий отфильтрованный и отсортированный набор.
- [ ] При пустом результате после фильтров показывается понятное сообщение и сохранение не выполняется.
- [ ] Регрессионные тесты GUI на save/filter/preset проходят.

---

## 6) Документация и фиксация этапа

- [x] `README.md` содержит актуальные примеры по CLI инструментам P1.
- [x] `docs/USER_GUIDE.md` синхронизирован по флагам и GUI вкладке `Инструменты`.
- [x] `docs/TECHNICAL.md` содержит ограничения инструментов.
- [x] `CHANGELOG.md` отражает все ключевые изменения P1.
- [ ] `docs/ROADMAP_P1_P3.md` содержит обновлённый статус P1.

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

### Осталось (до формального close P2)

- [ ] Финальный кросс-ОС ручной smoke/UX прогон на macOS/Linux и фиксация в release checklist.
- [ ] P2 closure-check выполнен в целевой ОС:
  - Unix: `./scripts/p2-closure-check.sh` (или `make p2-check`)
  - Windows: `.\scripts\p2-closure-check.ps1` (или `make p2-check-win`)

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
