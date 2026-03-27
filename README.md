# Network Scanner - Сканер локальной сети

[![GitHub](https://img.shields.io/github/license/RekadzeAV/network-scanner)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)

Кроссплатформенная утилита для сканирования локальных сетей с детальной аналитикой.

**Репозиторий:** https://github.com/RekadzeAV/network-scanner

## 🚀 Быстрый старт

### CLI версия (командная строка)

```bash
# Сборка
go build -o network-scanner ./cmd/network-scanner

# Запуск
./network-scanner
```

### GUI версия (графический интерфейс)

```bash
# Сборка
go build -o network-scanner-gui ./cmd/gui

# Запуск
./network-scanner-gui
```

## 📁 Структура проекта

```
Сканер локальной сети/
├── cmd/
│   ├── network-scanner/    # Точка входа CLI приложения
│   └── gui/                # Точка входа GUI приложения
├── internal/
│   ├── scanner/           # Логика сканирования
│   ├── network/           # Работа с сетью
│   ├── display/           # Отображение результатов (CLI)
│   └── gui/               # Компоненты графического интерфейса
├── docs/                  # Документация
├── scripts/               # Скрипты сборки
└── README.md             # Этот файл
```

## 📚 Документация

Полная документация находится в папке [docs/](docs/):

- **[Инструкция по эксплуатации](Инструкция%20по%20эксплуатации.md)** - Полная инструкция по эксплуатации программы (русский язык)
- **[README.md](docs/README.md)** - Основная документация с описанием возможностей
- **[USER_GUIDE.md](docs/USER_GUIDE.md)** - Подробное руководство пользователя
- **[GUI.md](docs/GUI.md)** - Документация по GUI версии приложения
- **[INSTALL.md](docs/INSTALL.md)** - Инструкции по установке
- **[QUICKSTART-macOS.md](docs/QUICKSTART-macOS.md)** - Быстрый старт для macOS
- **[TECHNICAL.md](docs/TECHNICAL.md)** - Техническая документация
- **[ARCHITECTURE.md](docs/ARCHITECTURE.md)** - Архитектура проекта
- **[ANALYSIS.md](docs/ANALYSIS.md)** - Анализ проекта
- **[RELEASE_SUMMARY_UI_RESULTS.md](docs/RELEASE_SUMMARY_UI_RESULTS.md)** - Краткий релиз-итог по UI результатов
- **[PR_DESCRIPTION_UI_RESULTS.md](docs/PR_DESCRIPTION_UI_RESULTS.md)** - Готовое описание PR по доработкам UI результатов
- **[RELEASE_ACCEPTANCE_CHECKLIST.md](docs/RELEASE_ACCEPTANCE_CHECKLIST.md)** - Финальный чеклист приемки перед релизом
- **[DEVELOPMENT_MAP.md](DEVELOPMENT_MAP.md)** - Детальная карта разработки проекта
- **[ROADMAP_P1_P3.md](docs/ROADMAP_P1_P3.md)** - Дорожная карта приоритетов P1–P3 (два этапа развития)
- **[CHANGELOG.md](CHANGELOG.md)** - История изменений проекта
- **[QUICKSTART_WINDOWS_BUILD.md](QUICKSTART_WINDOWS_BUILD.md)** - Быстрый старт: сборка для Windows на macOS
- **[RELEASE_NOTES_1.0.3.md](RELEASE_NOTES_1.0.3.md)** - Примечания к релизу 1.0.3

## 🔧 Сборка

### CLI версия

```bash
go build -o network-scanner ./cmd/network-scanner
```

### GUI версия

```bash
go build -o network-scanner-gui ./cmd/gui
```

### Использование скриптов

```bash
# macOS
./scripts/build-macos.sh

# Linux/Unix
./scripts/build.sh

# Windows (на Windows)
scripts\build.bat

# Сборка для Windows на macOS/Linux (кросскомпиляция)
./scripts/build-windows.sh  # Требует mingw-w64
```

### Smoke-проверки (регрессии CLI)

```bash
# Linux/macOS: базовый режим без топологии
./scripts/smoke-cli-no-topology.sh

# Linux/macOS: режим с топологией (проверка SNMP summary)
./scripts/smoke-cli-topology.sh
```

```powershell
# Windows PowerShell
.\scripts\smoke-cli-no-topology.ps1
.\scripts\smoke-cli-topology.ps1
```

Оба smoke-скрипта используют `127.0.0.1/32` и короткий диапазон портов для быстрого прогона.

### Smoke-проверка GUI

Для ручной проверки GUI-режимов используйте чеклист:
- [docs/GUI_SMOKE_CHECKLIST.md](docs/GUI_SMOKE_CHECKLIST.md)

## 📦 Требования

- Go 1.24 или выше
- Для GUI версии требуется C компилятор (GCC) из-за CGO
- Для кросскомпиляции в Windows на macOS/Linux требуется mingw-w64
- Для получения MAC адресов может потребоваться запуск с правами администратора

### Настройка для кросскомпиляции в Windows

Если вы хотите собирать Windows версию на macOS:

1. Установите mingw-w64: `brew install mingw-w64`
2. Проверьте окружение: `./scripts/setup-windows-env.sh`
3. Соберите: `./scripts/build-windows.sh`

Подробнее: [QUICKSTART_WINDOWS_BUILD.md](QUICKSTART_WINDOWS_BUILD.md) или [docs/SETUP_WINDOWS_CROSS_COMPILE.md](docs/SETUP_WINDOWS_CROSS_COMPILE.md)

## 🎯 Основные возможности

- 🔍 Автоматическое определение локальной сети
- 📡 Сканирование активных хостов
- 🔌 Сканирование портов TCP
- 🧭 Опциональный SNMP-опрос и построение топологии (`--topology`)
- 🗺️ Экспорт топологии в `json`, `graphml`, `png`, `svg` (для изображений нужен Graphviz `dot`)
- 🖥️ Определение типов устройств
- 📊 Аналитика по протоколам и портам
- 🏷️ Определение производителя по MAC адресу
- 🖼️ GUI-режимы результатов: `Таблица` / `Карточки` с сохранением выбора
- 📋 Табличный вывод с горизонтальной прокруткой и чипами портов
- 🧩 Карточки устройств с адаптивной сеткой (на узком экране 1 колонка)
- 🥧 Аналитика в GUI: 2 подтаблицы (табличный режим) или 2 круговые диаграммы (карточный режим)

## 🖥️ GUI: режимы отображения результатов

Во вкладке `Сканирование` доступны два режима:

- `Таблица`:
  - колонки `HostName`, `IP`, `MAC`, `Порты`;
  - порты отображаются как чипы с переносом на новую строку;
  - при нехватке ширины доступен горизонтальный скролл таблицы;
  - под таблицей выводится аналитика по протоколам и типам устройств.
- `Карточки`:
  - сетка карточек устройств (`HostName`, `IP`, `MAC`, порты-чипы);
  - на узком экране сетка переходит в одну колонку;
  - аналитика отображается двумя круговыми диаграммами.

Выбранный режим сохраняется в настройках приложения и автоматически восстанавливается при следующем запуске.
Дополнительно сохраняются сортировка, лимит чипов портов и быстрые фильтры результатов.

## 🧪 Новые CLI команды для топологии

```bash
# Базовое сканирование + построение топологии (вывод связей в консоль)
./network-scanner --topology

# Построение топологии и сохранение в JSON
./network-scanner --topology --output-format json --output-file topology.json

# Построение топологии и сохранение в GraphML
./network-scanner --topology --output-format graphml --output-file topology.graphml

# Построение топологии и экспорт в PNG (требуется Graphviz/dot)
./network-scanner --topology --output-format png --output-file topology.png

# Несколько SNMP community и увеличенный таймаут
./network-scanner --topology --snmp-community public,private,monitor --snmp-timeout 4
```

Ключевые флаги:
- `--topology` включает построение топологии после обычного сканирования.
- `--output-format` поддерживает `json`, `graphml`, `png`, `svg`.
- `--output-file` задает путь и имя файла для сохранения.
- `--snmp-community` принимает одну или несколько community-строк через запятую.
- `--snmp-timeout` задает SNMP-таймаут в секундах.

## 📝 Лицензия

Этот проект распространяется под лицензией MIT. См. файл [LICENSE](LICENSE) для подробностей.

## 🤝 Вклад в проект

Проект открыт для вклада! Если вы хотите улучшить проект:

1. Создайте форк репозитория
2. Создайте ветку для вашей функции (`git checkout -b feature/AmazingFeature`)
3. Зафиксируйте изменения (`git commit -m 'Add some AmazingFeature'`)
4. Отправьте в ветку (`git push origin feature/AmazingFeature`)
5. Откройте Pull Request

## 📋 История изменений

См. [CHANGELOG.md](CHANGELOG.md) для списка всех изменений в проекте.

## ⚠️ Предупреждение

Этот инструмент предназначен для использования только в ваших собственных сетях или сетях, где у вас есть явное разрешение на сканирование. Не используйте его для несанкционированного сканирования сетей.

## 🔗 Ссылки

- [Инструкция по эксплуатации](Инструкция%20по%20эксплуатации.md) - Полная инструкция по эксплуатации (русский язык)
- [Руководство пользователя](docs/USER_GUIDE.md) - Подробное руководство пользователя
- [Инструкция по установке](docs/INSTALL.md) - Инструкции по установке
- [Техническая документация](docs/TECHNICAL.md) - Техническая документация
- [Архитектура проекта](docs/ARCHITECTURE.md) - Архитектура проекта
- [Анализ проекта](docs/ANALYSIS.md) - Анализ проекта
- [Карта разработки](DEVELOPMENT_MAP.md) - Детальная карта разработки
- [История изменений](CHANGELOG.md) - История изменений
- [Быстрый старт: Windows сборка](QUICKSTART_WINDOWS_BUILD.md) - Сборка для Windows на macOS
