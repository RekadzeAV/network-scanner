# Быстрый старт для macOS

## 1. Установите Go (если еще не установлен)

```bash
# Через Homebrew
brew install go

# Или скачайте с https://go.dev/dl/
```

## 2. Соберите приложение

```bash
cd "Сканер локальной сети"
./scripts/build-macos.sh
```

Скрипт сохраняет бинарники в **`build/release/<YYYY-MM-DD-N>/`** в корне репозитория (подкаталог печатается при сборке; см. [BUILD_STRUCTURE.md](BUILD_STRUCTURE.md)).

## 3. Запустите

В примерах ниже замените `build/release/2026-04-24-1/` на фактический путь из вывода скрипта.

```bash
# Для Apple Silicon (M1/M2/M3)
./build/release/2026-04-24-1/network-scanner-darwin-arm64

# Для Intel Mac
./build/release/2026-04-24-1/network-scanner-darwin-amd64

# Или универсальный (если создан)
./build/release/2026-04-24-1/network-scanner-darwin-universal
```

## Пример использования

```bash
# Автоматическое определение сети
./build/release/2026-04-24-1/network-scanner-darwin-arm64

# Указать сеть вручную
./build/release/2026-04-24-1/network-scanner-darwin-arm64 --network 192.168.1.0/24

# Сканировать определенные порты
./build/release/2026-04-24-1/network-scanner-darwin-arm64 --ports 80,443,8080
```

Готово! 🎉

