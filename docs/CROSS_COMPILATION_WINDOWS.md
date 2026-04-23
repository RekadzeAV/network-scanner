# Анализ системы и компоненты для кроссплатформенной сборки на Windows

## 📋 Обзор

Данный документ содержит анализ текущей системы и полный список необходимых компонентов для выполнения кроссплатформенной сборки проекта Network Scanner на Windows для различных целевых платформ (Windows, Linux, macOS).

## 🔍 Анализ текущей системы

### Текущее состояние проекта

**Проект:** Network Scanner (GUI приложение)  
**Язык:** Go 1.24+  
**GUI Framework:** Fyne v2.7.1 (требует CGO)  
**Зависимости с CGO:**
- `fyne.io/fyne/v2` - требует CGO для работы
- `github.com/google/gopacket` - может использовать CGO для работы с libpcap

### Текущие скрипты сборки

**`scripts/build.bat`** - выполняет кроссплатформенную сборку для:
- Windows 64-bit (нативная)
- Linux 64-bit (кросскомпиляция)
- macOS Intel (кросскомпиляция)
- macOS Apple Silicon (кросскомпиляция)

**Проблема:** Скрипт использует простую кросскомпиляцию Go (`GOOS`/`GOARCH`), но при включенном CGO это не работает без дополнительных инструментов.

## ⚠️ Критические ограничения

### Проблема CGO при кросскомпиляции

Go поддерживает кросскомпиляцию **только когда CGO отключен**. При включенном CGO (что требуется для Fyne) необходимо:

1. **Для каждой целевой платформы нужен соответствующий кросс-компилятор**
2. **Для каждой целевой платформы нужны соответствующие системные библиотеки и заголовочные файлы**
3. **Настройка переменных окружения для указания путей к кросс-компиляторам**

### Особенности по платформам

| Платформа | Сложность | Требования |
|-----------|-----------|------------|
| **Windows** | ✅ Легко | MinGW-w64 GCC (уже установлен) |
| **Linux** | ⚠️ Средне | Кросс-компилятор для Linux (x86_64-linux-gnu-gcc или musl-gcc) |
| **macOS** | ❌ Очень сложно | macOS SDK + clang (практически невозможно легально на Windows) |

## 📦 Необходимые компоненты для кроссплатформенной сборки

### 1. Базовые компоненты (уже установлены)

#### ✅ Go (версия 1.24 или выше)
- **Статус:** Установлен (go1.25.5)
- **Проверка:** `go version`
- **Назначение:** Основной компилятор Go

#### ✅ C компилятор для Windows (MinGW-w64 GCC)
- **Статус:** Установлен (WinLibs MinGW-w64 GCC 15.2.0)
- **Проверка:** `gcc --version`
- **Назначение:** Компиляция для Windows (нативная платформа)
- **Путь:** Обычно `C:\mingw64\bin` или `C:\TDM-GCC-64\bin`

### 2. Компоненты для кросскомпиляции в Linux

#### 🔧 x86_64-linux-gnu-gcc (рекомендуется)

**Описание:** Кросс-компилятор GCC для Linux x86_64 на Windows

**Варианты установки:**

**Вариант 1: MSYS2 (рекомендуется)**
```powershell
# Установить MSYS2 (если еще не установлен)
# Скачать с https://www.msys2.org/

# В MSYS2 терминале:
pacman -S mingw-w64-x86_64-gcc
pacman -S mingw-w64-x86_64-toolchain

# Для Linux кросс-компиляции:
pacman -S mingw-w64-x86_64-gcc-linux-gnu
```

**Вариант 2: WinLibs (если доступен)**
```powershell
# Проверить наличие пакета для Linux кросс-компиляции
winget search linux-gcc
```

**Вариант 3: Вручную (сложно)**
- Скачать предкомпилированный кросс-компилятор
- Распаковать и добавить в PATH

**Проверка установки:**
```powershell
x86_64-linux-gnu-gcc --version
```

**Настройка для Go:**
```powershell
# Установить переменные окружения для кросскомпиляции Linux
$env:CC = "x86_64-linux-gnu-gcc"
$env:CXX = "x86_64-linux-gnu-g++"
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "1"
```

#### 🔧 Альтернатива: musl-gcc (для статической сборки)

**Описание:** Кросс-компилятор на основе musl libc для создания статических бинарников Linux

**Установка:**
```powershell
# Через MSYS2
pacman -S mingw-w64-x86_64-musl

# Или скачать с https://musl.cc/
```

**Преимущества:**
- Создает полностью статические бинарники
- Не требует системных библиотек на целевом Linux

**Настройка:**
```powershell
$env:CC = "x86_64-linux-musl-gcc"
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "1"
```

### 3. Компоненты для кросскомпиляции в macOS

#### ⚠️ macOS SDK + clang (практически невозможно на Windows)

**Проблема:** 
- macOS SDK доступен только на macOS (лицензионное ограничение Apple)
- Нельзя легально получить macOS SDK для использования на Windows
- Даже при наличии SDK, кросс-компиляция macOS на Windows крайне сложна

**Возможные решения:**

**Вариант 1: Использовать macOS для сборки (рекомендуется)**
- Собрать на реальной машине macOS
- Или использовать macOS в виртуальной машине (если есть лицензия)

**Вариант 2: Использовать CI/CD (GitHub Actions)**
- GitHub Actions предоставляет macOS runners
- Автоматическая сборка при коммитах

**Вариант 3: osxcross (не рекомендуется, нарушает лицензию)**
- Проект osxcross позволяет кросс-компилировать для macOS
- **⚠️ ВНИМАНИЕ:** Требует macOS SDK, который нельзя легально использовать на Windows
- Нарушает лицензионное соглашение Apple

**Рекомендация:** Не пытаться кросскомпилировать для macOS на Windows. Использовать macOS машину или CI/CD.

### 4. Дополнительные инструменты

#### 📝 Утилиты для проверки и отладки

**file (для проверки бинарников)**
```powershell
# Установить через MSYS2
pacman -S file

# Проверка собранного бинарника
file network-scanner-gui-linux-amd64
```

**objdump (для анализа бинарников)**
```powershell
# Обычно входит в MinGW-w64
objdump -f network-scanner-gui-linux-amd64
```

**ldd (для Linux, через WSL или Docker)**
```bash
# Проверить зависимости Linux бинарника
ldd network-scanner-gui-linux-amd64
```

## 🔧 Настройка окружения для кросскомпиляции

### Переменные окружения Go для кросскомпиляции

#### Для Linux (x86_64)
```powershell
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CC = "x86_64-linux-gnu-gcc"
$env:CXX = "x86_64-linux-gnu-g++"
$env:CGO_ENABLED = "1"
```

#### Для Linux (статическая сборка с musl)
```powershell
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CC = "x86_64-linux-musl-gcc"
$env:CGO_ENABLED = "1"
```

#### Для macOS (только на macOS)
```bash
export GOOS=darwin
export GOARCH=amd64  # или arm64 для Apple Silicon
export CC=clang
export CGO_ENABLED=1
```

### Скрипт для автоматической настройки

Создать файл `scripts/setup-cross-env.ps1`:
```powershell
# Настройка окружения для кросскомпиляции

function Set-LinuxCrossEnv {
    Write-Host "Настройка окружения для Linux кросскомпиляции..."
    
    # Проверка наличия кросс-компилятора
    $gcc = Get-Command x86_64-linux-gnu-gcc -ErrorAction SilentlyContinue
    if (-not $gcc) {
        Write-Error "x86_64-linux-gnu-gcc не найден. Установите через MSYS2: pacman -S mingw-w64-x86_64-gcc-linux-gnu"
        return $false
    }
    
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    $env:CC = "x86_64-linux-gnu-gcc"
    $env:CXX = "x86_64-linux-gnu-g++"
    $env:CGO_ENABLED = "1"
    
    Write-Host "✅ Окружение настроено для Linux кросскомпиляции"
    return $true
}

function Reset-Env {
    Write-Host "Сброс переменных окружения..."
    Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue
    Remove-Item Env:\CC -ErrorAction SilentlyContinue
    Remove-Item Env:\CXX -ErrorAction SilentlyContinue
    $env:CGO_ENABLED = "1"  # Оставляем CGO включенным
    Write-Host "✅ Переменные окружения сброшены"
}
```

## 📋 Итоговый чек-лист компонентов

### ✅ Обязательные компоненты (для Windows)

| Компонент | Статус | Команда проверки | Примечание |
|-----------|--------|------------------|------------|
| Go 1.24+ | ✅ Установлен | `go version` | Основной компилятор |
| MinGW-w64 GCC | ✅ Установлен | `gcc --version` | Для Windows сборки |
| CGO включен | ✅ | `go env CGO_ENABLED` | Должно быть `1` |

### ⚠️ Дополнительные компоненты (для Linux кросскомпиляции)

| Компонент | Статус | Команда проверки | Примечание |
|-----------|--------|------------------|------------|
| x86_64-linux-gnu-gcc | ❓ Требуется установка | `x86_64-linux-gnu-gcc --version` | Кросс-компилятор для Linux |
| Linux системные библиотеки | ❓ Требуется установка | - | Заголовочные файлы для Linux |
| MSYS2 (опционально) | ❓ Рекомендуется | `pacman --version` | Удобная установка кросс-компиляторов |

### ❌ Компоненты для macOS (не рекомендуется на Windows)

| Компонент | Статус | Альтернатива |
|-----------|--------|--------------|
| macOS SDK | ❌ Недоступен | Использовать macOS машину |
| clang для macOS | ❌ Недоступен | Использовать CI/CD (GitHub Actions) |

## 🚀 Рекомендуемый подход к кроссплатформенной сборке

### Вариант 1: Гибридный подход (рекомендуется)

1. **Windows:** Собирать на Windows (нативная сборка)
2. **Linux:** Собирать на Windows с кросс-компилятором (если установлен)
3. **macOS:** Собирать через CI/CD (GitHub Actions) или на macOS машине

### Вариант 2: Полностью через CI/CD

Использовать GitHub Actions для автоматической сборки всех платформ:
- Windows: `windows-latest` runner
- Linux: `ubuntu-latest` runner
- macOS: `macos-latest` runner

### Вариант 3: Docker для кросскомпиляции

Использовать Docker контейнеры с предустановленными кросс-компиляторами:
```dockerfile
FROM golang:1.24
RUN apt-get update && apt-get install -y gcc-x86-64-linux-gnu
```

## 📝 Обновление скрипта build.bat

Текущий скрипт `build.bat` не учитывает требования CGO. Необходимо обновить:

1. **Проверка наличия кросс-компиляторов** перед сборкой
2. **Установка переменных окружения** для каждой платформы
3. **Обработка ошибок** при отсутствии необходимых инструментов
4. **Пропуск macOS сборки** на Windows с предупреждением

## 🔍 Проверка готовности системы

### Скрипт проверки (check-cross-compilation.ps1)

```powershell
Write-Host "Проверка готовности к кроссплатформенной сборке..." -ForegroundColor Cyan

# Проверка Go
Write-Host "`n[1/5] Проверка Go..."
$goVersion = go version 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ $goVersion" -ForegroundColor Green
} else {
    Write-Host "❌ Go не установлен" -ForegroundColor Red
    exit 1
}

# Проверка GCC для Windows
Write-Host "`n[2/5] Проверка GCC для Windows..."
$gccVersion = gcc --version 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ GCC установлен" -ForegroundColor Green
    Write-Host "   $($gccVersion -split "`n" | Select-Object -First 1)"
} else {
    Write-Host "❌ GCC не найден" -ForegroundColor Red
    Write-Host "   Установите MinGW-w64 или TDM-GCC" -ForegroundColor Yellow
}

# Проверка CGO
Write-Host "`n[3/5] Проверка CGO..."
$cgoEnabled = go env CGO_ENABLED
if ($cgoEnabled -eq "1") {
    Write-Host "✅ CGO включен" -ForegroundColor Green
} else {
    Write-Host "⚠️  CGO отключен (требуется для Fyne)" -ForegroundColor Yellow
    Write-Host "   Установите: `$env:CGO_ENABLED = '1'" -ForegroundColor Yellow
}

# Проверка кросс-компилятора для Linux
Write-Host "`n[4/5] Проверка кросс-компилятора для Linux..."
$linuxGcc = Get-Command x86_64-linux-gnu-gcc -ErrorAction SilentlyContinue
if ($linuxGcc) {
    Write-Host "✅ x86_64-linux-gnu-gcc найден" -ForegroundColor Green
} else {
    Write-Host "⚠️  Кросс-компилятор для Linux не найден" -ForegroundColor Yellow
    Write-Host "   Установите через MSYS2: pacman -S mingw-w64-x86_64-gcc-linux-gnu" -ForegroundColor Yellow
    Write-Host "   Или используйте CI/CD для сборки Linux версии" -ForegroundColor Yellow
}

# Проверка инструментов для macOS
Write-Host "`n[5/5] Проверка инструментов для macOS..."
Write-Host "⚠️  Кросскомпиляция macOS на Windows невозможна" -ForegroundColor Yellow
Write-Host "   Используйте macOS машину или CI/CD (GitHub Actions)" -ForegroundColor Yellow

Write-Host "`n✅ Проверка завершена!" -ForegroundColor Cyan
```

## 📚 Дополнительные ресурсы

### Документация
- [Go CGO Documentation](https://pkg.go.dev/cmd/cgo)
- [Go Cross Compilation](https://go.dev/doc/install/source#crosscompile)
- [Fyne Cross Compilation](https://developer.fyne.io/started/cross-compiling)

### Инструменты
- [MSYS2](https://www.msys2.org/) - Среда для установки кросс-компиляторов
- [WinLibs](https://winlibs.com/) - Предкомпилированные GCC для Windows
- [musl.cc](https://musl.cc/) - Статические кросс-компиляторы

### CI/CD
- [GitHub Actions](https://github.com/features/actions) - Автоматическая сборка
- [GitLab CI](https://docs.gitlab.com/ee/ci/) - Альтернатива GitHub Actions

## 🎯 Выводы и рекомендации

### Краткий итог

1. **Для Windows:** Все необходимые компоненты уже установлены ✅
2. **Для Linux:** Требуется установка кросс-компилятора (x86_64-linux-gnu-gcc) ⚠️
3. **Для macOS:** Кросскомпиляция на Windows невозможна, использовать CI/CD или macOS машину ❌

### Рекомендации

1. **Приоритет 1:** Обновить скрипт `build.bat` с проверками и правильной настройкой окружения
2. **Приоритет 2:** Установить кросс-компилятор для Linux (если нужна локальная сборка)
3. **Приоритет 3:** Настроить CI/CD (GitHub Actions) для автоматической сборки всех платформ
4. **Приоритет 4:** Создать скрипт проверки готовности системы

---

**Версия документа:** 1.0.5  
**Дата создания:** 2024  
**Последнее обновление:** 2026-04-23


