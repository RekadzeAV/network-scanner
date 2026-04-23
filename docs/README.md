# Сканер локальной сети

Кроссплатформенная утилита для сканирования локальной сети с детальной аналитикой.

## Возможности

- 🔍 Автоматическое определение локальной сети
- 📡 Сканирование активных хостов
- 🔌 Сканирование портов TCP
- 🖥️ Определение типов устройств
- 📊 Аналитика по протоколам и портам
- 🏷️ Определение производителя по MAC адресу
- 📋 Красивый табличный вывод результатов
- 🧭 GUI-подрежимы результатов `Devices/Security` на вкладке сканирования
- 🧰 `Operations Center` с историей операций и действиями `Retry/Cancel`
- 🧾 `Host Details Drawer` с быстрыми действиями по выбранному хосту
- 🛡️ `Security Dashboard` с агрегированными findings и HTML-экспортом отчета

## Требования

- Go 1.24 или выше
- Для получения MAC адресов может потребоваться запуск с правами администратора (на некоторых системах)

> **Для macOS:** См. подробную инструкцию в [INSTALL.md](INSTALL.md)

## Быстрый старт

### CLI версия (командная строка)

```bash
# Установите зависимости
go mod download

# Соберите для текущей платформы
go build -o network-scanner ./cmd/network-scanner

# Запустите сканер (автоматически определит сеть)
./network-scanner
```

### GUI версия (графический интерфейс)

```bash
# Установите зависимости
go mod download

# Соберите GUI версию
go build -o network-scanner-gui ./cmd/gui

# Запустите GUI приложение
./network-scanner-gui
```

### Smoke-проверка адаптивности GUI (разрешение/DPI)

```bash
# Linux/macOS
./scripts/smoke-gui-resolution.sh ./network-scanner-gui
```

```powershell
# Windows PowerShell
.\scripts\smoke-gui-resolution.ps1 -GuiExe .\network-scanner-gui.exe
```

Скрипты запускают GUI и печатают матрицу ручной проверки (`1366x768` ... `4K`) и критерии приемки для оконного и полноэкранного режимов.

## Установка

### Сборка из исходников

```bash
# Перейдите в директорию проекта
cd "Сканер локальной сети"

# Установите зависимости
go mod download

# Соберите для текущей платформы
go build -o network-scanner

# Или используйте скрипты сборки:
# macOS (рекомендуется)
chmod +x scripts/build-macos.sh
./scripts/build-macos.sh

# Linux/macOS (все платформы)
chmod +x scripts/build.sh
./scripts/build.sh

# Windows (из cmd/PowerShell в корне репозитория)
# scripts\build.bat

# Или соберите для других платформ:
# Linux 64-bit
GOOS=linux GOARCH=amd64 go build -o network-scanner-linux-amd64

# Windows 64-bit
GOOS=windows GOARCH=amd64 go build -o network-scanner-windows-amd64.exe

# macOS 64-bit (Intel)
GOOS=darwin GOARCH=amd64 go build -o network-scanner-darwin-amd64

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o network-scanner-darwin-arm64
```

Релизные скрипты (`scripts/build.sh`, `scripts/build-macos.sh`, `scripts/build.bat` и др.) складывают готовые бинарники в **`build/release/`** в корне репозитория (см. [BUILD_STRUCTURE.md](BUILD_STRUCTURE.md)).

## Использование

### Базовое использование

```bash
# Автоматическое определение сети
./network-scanner

# Указание сети вручную
./network-scanner --network 192.168.1.0/24

# Сканирование определенных портов
./network-scanner --ports 80,443,8080

# Сканирование диапазона портов
./network-scanner --ports 1-1000

# Настройка таймаута
./network-scanner --timeout 5

# Настройка количества потоков
./network-scanner --threads 200
```

### Параметры командной строки

- `--network` - Диапазон сети для сканирования (например: `192.168.1.0/24`)
- `--timeout` - Таймаут сканирования в секундах
- `--ports` - Диапазон портов для сканирования
  - Можно указать список: `80,443,8080`
  - Или диапазон: `1-1000`
  - Или комбинацию: `80,443,8080-8090`
- `--threads` - Количество потоков для сканирования
- `--show-closed` - Показывать закрытые порты
- `--udp` - Включить проверку популярных UDP-портов

## Примеры вывода

### Результаты сканирования

Утилита выводит таблицу с информацией о каждом обнаруженном устройстве:
- IP адрес
- MAC адрес
- Hostname
- Открытые порты с сервисами
- Протоколы
- Тип устройства
- Производитель

### Аналитика

После сканирования выводится детальная аналитика:
- Статистика по протоколам в сети
- Используемые порты и их назначение
- Типы устройств
- Общая статистика

## Особенности

### Определение MAC адресов

Для получения MAC адресов утилита использует ARP запросы. На некоторых системах это может требовать прав администратора. Если MAC адреса не определяются, это нормально - остальная функциональность будет работать.

### Определение типов устройств

Тип устройства определяется на основе:
- Открытых портов
- Протоколов
- MAC адреса (OUI)

### Производительность

Утилита использует многопоточное сканирование для ускорения процесса. Количество потоков можно настроить через параметр `--threads`.

## Ограничения

- По умолчанию сканируются TCP-порты; UDP-проверка включается отдельно через `--udp`
- MAC адреса могут не определяться без прав администратора
- Некоторые устройства могут не отвечать на ping/ARP запросы

## Лицензия

Этот проект создан для образовательных целей.

## Документация

Проект включает подробную документацию:

### Smoke и closure проверки (оперативно)

```bash
# Linux/macOS
./scripts/smoke-cli-tools.sh
./scripts/stage2-p1-closure-check.sh
```

```powershell
# Windows PowerShell
.\scripts\smoke-cli-tools.ps1
.\scripts\stage2-p1-closure-check.ps1
```

- `smoke-cli-tools` проверяет tool-режимы `--ping`/`--dns` и включает детерминированную проверку CLI whois-пути через RDAP fallback (`go test ./cmd/network-scanner -run WhoisUsesRDAPFallback`).
- `stage2-p1-closure-check` дополнительно включает `go test ./cmd/network-scanner -run Whois`, чтобы регрессии в `runToolsMode` для `--whois` ловились на этапе формального closure.

### Операционный индекс (release/closure)

- **[RELEASE_ACCEPTANCE_CHECKLIST.md](RELEASE_ACCEPTANCE_CHECKLIST.md)** - Финальный чеклист приемки перед релизом
- **[P1_CLOSURE_CHECKLIST.md](P1_CLOSURE_CHECKLIST.md)** - Формальное закрытие Stage 1 / P1
- **[P3_CLOSURE_CHECKLIST.md](P3_CLOSURE_CHECKLIST.md)** - Формальное закрытие Stage 1 / P3 и CI sign-off
- **[GUI_SMOKE_CHECKLIST.md](GUI_SMOKE_CHECKLIST.md)** - Ручной smoke-чеклист GUI
- **[RELEASE_READINESS_SNAPSHOT.md](RELEASE_READINESS_SNAPSHOT.md)** - Текущий снимок готовности релиза
- **[CHECKLIST_STATUS_INDEX.md](CHECKLIST_STATUS_INDEX.md)** - Единый индекс статусов всех checklist-документов
- **[RELEASE_READY_GAP_LIST.md](RELEASE_READY_GAP_LIST.md)** - Короткий backlog оставшихся задач до финального sign-off
- **[P0_SIGNOFF_RUNBOOK.md](P0_SIGNOFF_RUNBOOK.md)** - Пошаговый runbook закрытия блокирующего P0 (Cross-OS + CI evidence)
- **[RELEASE_OPERATIONS_CHEATSHEET.md](RELEASE_OPERATIONS_CHEATSHEET.md)** - Команды релизного прогона и расположение локальных бинарников (`build/release/`)
- **[BUILD_STRUCTURE.md](BUILD_STRUCTURE.md)** - Структура каталогов релизной сборки
- **[STAGE2_100_COMMIT_READY.md](STAGE2_100_COMMIT_READY.md)** - Краткий commit-ready итог по Stage2 и remaining шагам sign-off
- **[COMMIT_READY_STAGE2_SIGNOFF.md](COMMIT_READY_STAGE2_SIGNOFF.md)** - Готовые варианты commit message для пакета Stage2 sign-off
- **[ROADMAP_P1_P3.md](ROADMAP_P1_P3.md)** - Канонический статус фаз Stage 1/2 (`P1..P3`)
- **[DETAILED_BACKLOG_P3_STAGE2.md](DETAILED_BACKLOG_P3_STAGE2.md)** - Детализированный backlog Stage 1 P3 + Stage 2 P1/P2/P3
- **[RELEASE_READINESS_PR_READY.md](RELEASE_READINESS_PR_READY.md)** - Готовые short/long блоки статуса для PR
- **[DOCS_SYNC_SUMMARY_2026-04-23.md](DOCS_SYNC_SUMMARY_2026-04-23.md)** - Сводка синхронизации документации
- **[DOCS_SYNC_PR_SNIPPET_2026-04-23.md](DOCS_SYNC_PR_SNIPPET_2026-04-23.md)** - Короткий RU блок для PR-комментария
- **[DOCS_SYNC_PR_SNIPPET_2026-04-23_EN.md](DOCS_SYNC_PR_SNIPPET_2026-04-23_EN.md)** - Короткий EN блок для PR-комментария
- **[FINAL_PR_COMMENT_READY.md](FINAL_PR_COMMENT_READY.md)** - Финальный ready-to-paste комментарий в PR
- **[MANUAL_SIGNOFF_TEMPLATE.md](MANUAL_SIGNOFF_TEMPLATE.md)** - Шаблон ручного sign-off
- **[MANUAL_SIGNOFF_DRAFT.md](MANUAL_SIGNOFF_DRAFT.md)** - Черновик sign-off с предзаполненными auto-evidence

### Быстрый preflight перед финальным sign-off (Windows)

```powershell
.\scripts\p0-signoff-preflight.ps1
# или
make p0-preflight-win
```

Preflight проверяет критические блокеры (`GITHUB_TOKEN`, рабочий `bash/sh` runtime для Unix closure, наличие успешного recent `ci.yml` run).

### Единый агрегированный статус Stage2 sign-off (Windows)

```powershell
.\scripts\stage2-signoff-status.ps1
# или
make stage2-signoff-status-win
```

### Docs sanity: проверка локальных markdown-ссылок (Windows)

```powershell
.\scripts\docs-link-check.ps1
# или
make docs-link-check-win
```

- **[Инструкция по эксплуатации](../Инструкция%20по%20эксплуатации.md)** - Полная инструкция по эксплуатации программы (русский язык)
- **[README.md](../README.md)** - Основная документация проекта
- **[README.md](README.md)** - Основная документация (этот файл)
- **[USER_GUIDE.md](USER_GUIDE.md)** - Подробное руководство пользователя с примерами
- **[GUI.md](GUI.md)** - Документация по GUI версии приложения
- **[INSTALL.md](INSTALL.md)** - Инструкции по установке для разных платформ
- **[QUICKSTART-macOS.md](QUICKSTART-macOS.md)** - Быстрый старт для macOS
- **[TECHNICAL.md](TECHNICAL.md)** - Техническая документация для разработчиков
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Описание архитектуры проекта
- **[ANALYSIS.md](ANALYSIS.md)** - Анализ реализации и рекомендации
- **[DEVELOPMENT_MAP.md](../DEVELOPMENT_MAP.md)** - Детальная карта разработки проекта
- **[CHANGELOG.md](../CHANGELOG.md)** - История изменений проекта
- **[QUICKSTART_WINDOWS_BUILD.md](../QUICKSTART_WINDOWS_BUILD.md)** - Быстрый старт: сборка для Windows на macOS
- **[RELEASE_NOTES_1.0.3.md](../RELEASE_NOTES_1.0.3.md)** - Примечания к релизу 1.0.3
- **[UI_IMPLEMENTATION_BACKLOG.md](UI_IMPLEMENTATION_BACKLOG.md)** - Детализированный и актуализированный backlog UI-рефакторинга (`P0..P5`)
- **[PR_DESCRIPTION_UI_RESULTS.md](PR_DESCRIPTION_UI_RESULTS.md)** - Актуальный шаблон описания PR по изменениям GUI/результатов

## Поддержка

При возникновении проблем:
1. Убедитесь, что у вас установлена актуальная версия Go
2. Проверьте, что вы находитесь в локальной сети
3. На некоторых системах может потребоваться запуск с правами администратора для получения MAC адресов
4. См. [USER_GUIDE.md](USER_GUIDE.md) для подробной информации

