# Быстрый старт: Сборка под Windows на macOS

## Шаг 1: Установка Homebrew (если еще не установлен)

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

## Шаг 2: Установка mingw-w64

```bash
brew install mingw-w64
```

## Шаг 3: Проверка окружения

```bash
./scripts/setup-windows-env.sh
```

## Шаг 4: Сборка

```bash
./scripts/build-windows.sh
```

Собранный файл будет в `release/YYYY-MM-DD/network-scanner-gui-windows-amd64.exe`

## Подробная документация

См. [docs/SETUP_WINDOWS_CROSS_COMPILE.md](docs/SETUP_WINDOWS_CROSS_COMPILE.md) для подробных инструкций и решения проблем.

