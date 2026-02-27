# Структура сборки релизов

## Обзор

Начиная с версии 1.0.2, каждый релиз создает отдельные директории для каждой операционной системы. Это упрощает распространение и использование приложения.

## Структура директорий

```
release/
└── YYYY-MM-DD-N[-debug]/
    ├── windows/
    │   ├── network-scanner.exe
    │   ├── network-scanner-gui.exe
    │   ├── README.md
    │   └── Инструкция по эксплуатации.md
    ├── linux/
    │   ├── network-scanner
    │   ├── network-scanner-gui
    │   ├── README.md
    │   └── Инструкция по эксплуатации.md
    ├── darwin-amd64/
    │   ├── network-scanner
    │   ├── network-scanner-gui
    │   ├── README.md
    │   └── Инструкция по эксплуатации.md
    ├── darwin-arm64/
    │   ├── network-scanner
    │   ├── network-scanner-gui
    │   ├── README.md
    │   └── Инструкция по эксплуатации.md
    └── Инструкция по эксплуатации.md
```

## Платформы

### Windows
- **Директория:** `windows/`
- **Файлы:**
  - `network-scanner.exe` - CLI версия
  - `network-scanner-gui.exe` - GUI версия
- **Архитектура:** amd64 (64-bit)

### Linux
- **Директория:** `linux/`
- **Файлы:**
  - `network-scanner` - CLI версия
  - `network-scanner-gui` - GUI версия
- **Архитектура:** amd64 (64-bit)
- **Примечание:** После скачивания нужно сделать файлы исполняемыми: `chmod +x network-scanner*`

### macOS Intel
- **Директория:** `darwin-amd64/`
- **Файлы:**
  - `network-scanner` - CLI версия
  - `network-scanner-gui` - GUI версия
- **Архитектура:** amd64 (Intel Mac)

### macOS Apple Silicon
- **Директория:** `darwin-arm64/`
- **Файлы:**
  - `network-scanner` - CLI версия
  - `network-scanner-gui` - GUI версия
- **Архитектура:** arm64 (M1, M2, M3, и т.д.)

## Сборка

### Релизная версия (без логирования)
```bash
scripts\build.bat release
# или просто
scripts\build.bat
```

### Debug версия (с логированием)
```bash
scripts\build.bat debug
```

## Распространение

Для распространения можно:

1. **Создать архив для каждой платформы:**
   - `network-scanner-windows.zip` - содержимое `windows/`
   - `network-scanner-linux.tar.gz` - содержимое `linux/`
   - `network-scanner-macos-intel.zip` - содержимое `darwin-amd64/`
   - `network-scanner-macos-apple-silicon.zip` - содержимое `darwin-arm64/`

2. **Или распространять всю директорию релиза:**
   - Пользователи могут выбрать нужную поддиректорию для своей ОС

## Преимущества новой структуры

1. **Удобство:** Каждая платформа в своей директории
2. **Чистота:** Нет смешивания файлов разных платформ
3. **Документация:** README для каждой платформы с инструкциями
4. **Распространение:** Легко создать отдельные архивы для каждой ОС
5. **Организация:** Понятная структура для пользователей

## Скрипты сборки

- `scripts/build.bat` - Основной скрипт (обновлен для новой структуры)
- `scripts/build-os-separate.bat` - Альтернативный скрипт с расширенным выводом

Оба скрипта создают одинаковую структуру директорий.
