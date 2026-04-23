# Требования для локальной сборки на Windows

## Анализ требований

### ✅ Обязательные компоненты

#### 1. Go (версия 1.24 или выше)
- **Статус**: Обязательно
- **Проверка установки**: 
  ```powershell
  go version
  ```
- **Установка**: Скачать с https://go.dev/dl/
- **Примечание**: В документации указано, что уже установлена версия go1.25.5

#### 2. C компилятор (GCC)
- **Статус**: Обязательно для GUI версии
- **Причина**: Fyne framework требует CGO для работы на Windows
- **Проверка установки**: 
  ```powershell
  gcc --version
  ```
- **Варианты установки**:

  **Вариант 1: TDM-GCC (рекомендуется)**
  - Сайт: https://jmeubank.github.io/tdm-gcc/
  - Простой установщик, автоматически добавляет в PATH
  - Перезапуск терминала после установки

  **Вариант 2: MinGW-w64 через MSYS2**
  - Сайт: https://www.msys2.org/
  - Команды установки:
    ```bash
    pacman -S mingw-w64-x86_64-gcc
    pacman -S mingw-w64-x86_64-toolchain
    ```
  - Добавить `C:\msys64\mingw64\bin` в PATH

  **Вариант 3: Chocolatey (если установлен)**
  ```powershell
  choco install mingw -y
  ```

#### 3. CGO должен быть включен
- **Статус**: Обязательно для GUI версии
- **Проверка**: 
  ```powershell
  go env CGO_ENABLED
  ```
- **Должно быть**: `1`
- **Установка** (если `0`):
  ```powershell
  $env:CGO_ENABLED = "1"
  ```

### 📦 Зависимости проекта

Зависимости автоматически управляются через Go modules (`go.mod`):

#### Основные зависимости:
- **fyne.io/fyne/v2 v2.7.1** - GUI framework (требует CGO)
- **github.com/google/gopacket v1.1.19** - работа с пакетами сети
- **github.com/jedib0t/go-pretty/v6 v6.5.4** - форматирование вывода

#### Установка зависимостей:
```powershell
go mod download
```

### 🔧 Процесс сборки

#### CLI версия (командная строка)

**Базовая сборка:**
```powershell
# Перейти в директорию проекта
cd "d:\Разработка через ИИ\network-scanner"

# Установить зависимости
go mod download

# Собрать CLI версию
go build -o network-scanner.exe ./cmd/network-scanner
```

**Запуск после сборки:**
```powershell
.\network-scanner.exe
# Или с параметрами:
.\network-scanner.exe --network 192.168.1.0/24 --ports 80,443,8080
```

#### GUI версия (графический интерфейс)

**Базовая сборка:**
```powershell
# Перейти в директорию проекта
cd "d:\Разработка через ИИ\network-scanner"

# Установить зависимости
go mod download

# Собрать GUI версию (требует CGO и GCC)
go build -o network-scanner-gui.exe ./cmd/gui
```

**Использование скрипта сборки:**
```powershell
.\scripts\build.bat
```
✅ **Примечание**: Скрипт `build.bat` собирает обе версии (CLI и GUI) для всех поддерживаемых платформ. Готовые файлы попадают в **`build\release\<YYYY-MM-DD-N>\`** (например **`windows\network-scanner.exe`**); см. [BUILD_STRUCTURE.md](BUILD_STRUCTURE.md).

**Запуск после сборки:**
```powershell
.\network-scanner-gui.exe
```

### 🔍 Проверка готовности к сборке

**Чек-лист перед сборкой:**

```powershell
# 1. Проверить Go
go version
# Должно показать: go version go1.24.x или выше

# 2. Проверить GCC
gcc --version
# Должно показать версию GCC (не ошибку)

# 3. Проверить CGO
go env CGO_ENABLED
# Должно быть: CGO_ENABLED="1"

# 4. Проверить, что в PATH есть GCC
$env:PATH -split ';' | Select-String -Pattern 'gcc|mingw|tdm'
# Должен найти путь к GCC

# 5. Проверить зависимости
go mod download
# Должно завершиться без ошибок
```

### 🚨 Частые проблемы и решения

#### Ошибка: "gcc: executable file not found"
**Причина**: GCC не установлен или не в PATH  
**Решение**: 
```powershell
# Проверить PATH
$env:PATH -split ';' | Select-String -Pattern 'gcc|mingw|tdm'

# Добавить путь к GCC (пример для TDM-GCC)
$env:PATH += ";C:\TDM-GCC-64\bin"
```

#### Ошибка: "CGO_ENABLED=0"
**Причина**: CGO отключен  
**Решение**:
```powershell
$env:CGO_ENABLED = "1"
# Или установить глобально через go env
```

#### Ошибка: "package cmd/network-scanner is not in GOROOT or GOPATH"
**Причина**: Неправильный путь к проекту или отсутствие файла  
**Решение**: Убедитесь, что вы находитесь в корне проекта и файл `cmd/network-scanner/main.go` существует:
```powershell
# Проверить структуру
ls cmd/network-scanner/main.go
# Если файл существует, попробовать сборку:
go build -o network-scanner.exe ./cmd/network-scanner
```

#### Ошибка с правами доступа
**Причина**: Для работы с сетевыми интерфейсами нужны права администратора  
**Решение**: Запускать от имени администратора (правой кнопкой → "Запуск от имени администратора")

### 📋 Итоговый список требований

| Компонент | Статус | Версия/Требования | Проверка |
|-----------|--------|-------------------|----------|
| Go | ✅ Обязательно | 1.24+ | `go version` |
| GCC | ✅ Обязательно | Любая современная | `gcc --version` |
| CGO | ✅ Обязательно | Включен (1) | `go env CGO_ENABLED` |
| Зависимости | ✅ Автоматически | Управляются через go.mod | `go mod download` |
| Права админа | ⚠️ Рекомендуется | Для полной функциональности | - |

### 🔄 Рекомендации по улучшению

1. ✅ **Скрипт build.bat** - работает корректно, собирает обе версии
2. ✅ **Документация** - актуализирована, содержит информацию о CLI и GUI версиях
3. ⚠️ **Добавить проверки в скрипт** - можно добавить проверку наличия GCC и CGO перед сборкой GUI
4. ✅ **Обе версии доступны** - CLI и GUI версии полностью функциональны

### 📝 Минимальные команды для быстрой сборки

**CLI версия (не требует CGO/GCC):**
```powershell
# 1. Установить зависимости
go mod download

# 2. Собрать CLI версию
go build -o network-scanner.exe ./cmd/network-scanner

# 3. Запустить
.\network-scanner.exe
```

**GUI версия (требует CGO и GCC):**
```powershell
# 1. Установить зависимости
go mod download

# 2. Собрать GUI версию
go build -o network-scanner-gui.exe ./cmd/gui

# 3. Запустить
.\network-scanner-gui.exe
```

**Обе версии через скрипт:**
```powershell
.\scripts\build.bat
```

### 🌐 Альтернативные варианты сборки

Если установка C компилятора проблематична:

1. **Кросскомпиляция из Linux/macOS** - собрать на другой платформе с настройкой `GOOS=windows`
2. **Docker** - использовать Docker контейнер для сборки
3. **CI/CD** - использовать GitHub Actions или другой CI/CD для автоматической сборки
4. **GitHub Releases** - использовать готовые бинарники из релизов




