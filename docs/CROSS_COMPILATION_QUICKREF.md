# Краткая справка: Компоненты для кроссплатформенной сборки

## 📋 Быстрая таблица компонентов

| Компонент | Windows | Linux | macOS | Статус | Команда проверки |
|-----------|---------|-------|-------|--------|------------------|
| **Go 1.24+** | ✅ Обязательно | ✅ Обязательно | ✅ Обязательно | ✅ Установлен | `go version` |
| **GCC (MinGW-w64)** | ✅ Обязательно | ❌ | ❌ | ✅ Установлен | `gcc --version` |
| **CGO включен** | ✅ Обязательно | ✅ Обязательно | ✅ Обязательно | ✅ | `go env CGO_ENABLED` |
| **x86_64-linux-gnu-gcc** | ⚠️ Для Linux | ❌ | ❌ | ❓ Требуется | `x86_64-linux-gnu-gcc --version` |
| **macOS SDK + clang** | ❌ Невозможно | ❌ | ✅ Обязательно | ❌ | - |

## 🎯 Что нужно для каждой платформы

### ✅ Windows (нативная сборка)
- Go 1.24+ ✅
- MinGW-w64 GCC ✅
- CGO включен ✅

**Статус:** Готово к сборке

### ⚠️ Linux (кросскомпиляция)
- Go 1.24+ ✅
- x86_64-linux-gnu-gcc ❓
- Linux системные библиотеки ❓

**Статус:** Требуется установка кросс-компилятора

**Установка:**
```powershell
# Через MSYS2
pacman -S mingw-w64-x86_64-gcc-linux-gnu
```

### ❌ macOS (кросскомпиляция)
- Go 1.24+ ✅
- macOS SDK ❌ (недоступен на Windows)
- clang для macOS ❌

**Статус:** Невозможно на Windows

**Решение:** Использовать macOS машину или CI/CD

## 🔧 Переменные окружения для сборки

### Windows
```powershell
# Обычно не требуется (нативная платформа)
go build -o network-scanner-gui.exe ./cmd/gui
```

### Linux
```powershell
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CC = "x86_64-linux-gnu-gcc"
$env:CGO_ENABLED = "1"
go build -o network-scanner-gui-linux-amd64 ./cmd/gui
```

### macOS (только на macOS)
```bash
export GOOS=darwin
export GOARCH=amd64  # или arm64
export CC=clang
export CGO_ENABLED=1
go build -o network-scanner-gui-darwin-amd64 ./cmd/gui
```

## 📝 Быстрая проверка

Запустите скрипт проверки:
```powershell
.\scripts\check-cross-compilation.ps1
```

## 📚 Подробная документация

- **Полный анализ:** `docs/CROSS_COMPILATION_WINDOWS.md`
- **Требования Windows:** `docs/BUILD_REQUIREMENTS_WINDOWS.md`
- **Установка:** `docs/INSTALL_WINDOWS.md`

## ⚡ Рекомендации

1. **Для Windows:** Все готово ✅
2. **Для Linux:** Установите кросс-компилятор или используйте CI/CD
3. **Для macOS:** Используйте CI/CD (GitHub Actions) или macOS машину

---

**Последнее обновление:** 2026-04-23


