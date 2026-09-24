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

### Актуальные документы

- **[README.md](../README.md)** — основная документация проекта (корень репозитория)
- **[PROJECT_PROMPT.md](PROJECT_PROMPT.md)** — единый промт/технический контекст проекта для продолжения разработки
- **[USER_GUIDE.md](USER_GUIDE.md)** — подробное руководство пользователя с примерами
- **[GUI.md](GUI.md)** — документация по GUI-версии приложения
- **[ARCHITECTURE.md](ARCHITECTURE.md)** — описание архитектуры проекта
- **[TECHNICAL.md](TECHNICAL.md)** — техническая документация для разработчиков
- **[PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md)** — структура проекта
- **[ROADMAP.md](ROADMAP.md)** — канонический roadmap
- **[ROADMAP.md](ROADMAP.md)** — roadmap и план реализации v2.3
- **[UNIFIED_OPTIMIZED_PLAN_2026-09-15.md](UNIFIED_OPTIMIZED_PLAN_2026-09-15.md)** — единый оптимизированный план работ
- **[THREE_PLANS_ANALYSIS_2026-09-15.md](THREE_PLANS_ANALYSIS_2026-09-15.md)** — три плана анализа проекта
- **[INSTALL.md](INSTALL.md)** — инструкции по установке для разных платформ
- **[QUICKSTART-macOS.md](QUICKSTART-macOS.md)** — быстрый старт для macOS
- **[BUILD_STRUCTURE.md](BUILD_STRUCTURE.md)** — структура каталогов релизной сборки (`build/release/`)
- **[LOGGING.md](LOGGING.md)** — система логирования
- **[deployment.md](deployment.md)** — развертывание
- **[GRAPHML_COMPATIBILITY_CHECK.md](GRAPHML_COMPATIBILITY_CHECK.md)** — ручная проверка совместимости GraphML (yEd/Gephi)
- **[swagger.yaml](swagger.yaml)** — спецификация REST API (статус: требует regenerate под v2.3)
- **[CHANGELOG.md](../CHANGELOG.md)** — история изменений

### Сборка и кросс-компиляция

- [BUILD_REQUIREMENTS_WINDOWS.md](BUILD_REQUIREMENTS_WINDOWS.md), [BUILD_STRUCTURE.md](BUILD_STRUCTURE.md)
- [CROSS_COMPILATION_WINDOWS.md](CROSS_COMPILATION_WINDOWS.md), [CROSS_COMPILATION_QUICKREF.md](CROSS_COMPILATION_QUICKREF.md)
- [INSTALL_WINDOWS.md](INSTALL_WINDOWS.md), [INSTALL_LINUX_CROSS_COMPILER.md](INSTALL_LINUX_CROSS_COMPILER.md)
- [SETUP_WINDOWS_CROSS_COMPILE.md](SETUP_WINDOWS_CROSS_COMPILE.md), [GIT_SETUP.md](GIT_SETUP.md)

### Smoke и preflight проверки

```bash
# Linux/macOS
./scripts/smoke-cli-tools.sh
./scripts/smoke-gui-resolution.sh ./network-scanner-gui
```

```powershell
# Windows PowerShell
.\scripts\smoke-cli-tools.ps1
.\scripts\smoke-gui-resolution.ps1 -GuiExe .\network-scanner-gui.exe
.\scripts\docs-link-check.ps1   # или make docs-link-check-win
```

### Архив

Завершённые и устаревшие документы — в [archive/](archive/):

- `archive/2026-01-release-cycle/` — релизный цикл 1.0.x
- `archive/2026-04-docs-sync/` — синхронизация документации 2026-04
- `archive/2026-08-ui-tests/` — цикл UI-тестов 2026-08
- `archive/2026-09-audit/` — аудит и closure-чеклисты v2.0–v2.2
- `archive/2026-09-15-docs-sync/` — планы и отчёты, заменённые циклом 2026-09-15

## Поддержка

При возникновении проблем:

1. Убедитесь, что у вас установлена актуальная версия Go
2. Проверьте, что вы находитесь в локальной сети
3. На некоторых системах может потребоваться запуск с правами администратора для получения MAC адресов
4. См. [USER_GUIDE.md](USER_GUIDE.md) для подробной информации
