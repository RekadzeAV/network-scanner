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
./build-macos.sh
```

## 3. Запустите

```bash
# Для Apple Silicon (M1/M2/M3)
./dist/network-scanner-darwin-arm64

# Для Intel Mac
./dist/network-scanner-darwin-amd64

# Или универсальный (если создан)
./dist/network-scanner-darwin-universal
```

## Пример использования

```bash
# Автоматическое определение сети
./dist/network-scanner-darwin-arm64

# Указать сеть вручную
./dist/network-scanner-darwin-arm64 --network 192.168.1.0/24

# Сканировать определенные порты
./dist/network-scanner-darwin-arm64 --ports 80,443,8080
```

Готово! 🎉

