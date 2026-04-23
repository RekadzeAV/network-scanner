@echo off
REM Скрипт для сборки сканера сети с отдельными директориями для каждой ОС
REM Использование: build-os-separate.bat [release|debug]
REM   release - релизная версия без лога (версия 1, по умолчанию)
REM   debug   - тестовая версия с логом (версия 2)

set BUILD_TYPE=%1
if "%BUILD_TYPE%"=="" set BUILD_TYPE=release
if /i "%BUILD_TYPE%"=="release" (
    set BUILD_VERSION=1
    set BUILD_TAGS=
    set BUILD_SUFFIX=
    echo Сборка РЕЛИЗНОЙ версии (без лога) - версия %BUILD_VERSION%
) else if /i "%BUILD_TYPE%"=="debug" (
    set BUILD_VERSION=2
    set BUILD_TAGS=-tags debug
    set BUILD_SUFFIX=-debug
    echo Сборка ТЕСТОВОЙ версии (с логом) - версия %BUILD_VERSION%
) else (
    echo Ошибка: неверный тип сборки. Используйте: release или debug
    exit /b 1
)

echo.
echo Сборка Network Scanner v1.0.%BUILD_VERSION% (%BUILD_TYPE%)...
echo.

REM Создаем директорию для бинарников с датой сборки и номером
for /f "usebackq tokens=*" %%i in (`powershell -Command "Get-Date -Format 'yyyy-MM-dd'"`) do set BUILD_DATE=%%i

REM Находим следующий доступный номер сборки
set BUILD_NUM=1
:find_build_num
set RELEASE_DIR=build\release\%BUILD_DATE%-%BUILD_NUM%%BUILD_SUFFIX%
if exist "%RELEASE_DIR%" (
    set /a BUILD_NUM+=1
    goto find_build_num
)

REM Создаем основную директорию релиза
if not exist "%RELEASE_DIR%" mkdir "%RELEASE_DIR%"

REM Создаем поддиректории для каждой ОС
set WINDOWS_DIR=%RELEASE_DIR%\windows
set LINUX_DIR=%RELEASE_DIR%\linux
set DARWIN_AMD64_DIR=%RELEASE_DIR%\darwin-amd64
set DARWIN_ARM64_DIR=%RELEASE_DIR%\darwin-arm64

mkdir "%WINDOWS_DIR%" 2>nul
mkdir "%LINUX_DIR%" 2>nul
mkdir "%DARWIN_AMD64_DIR%" 2>nul
mkdir "%DARWIN_ARM64_DIR%" 2>nul

echo 📦 Бинарники будут сохранены в: %RELEASE_DIR%\ (сборка #%BUILD_NUM%, версия %BUILD_VERSION%)
echo    - Windows: %WINDOWS_DIR%\
echo    - Linux: %LINUX_DIR%\
echo    - macOS Intel: %DARWIN_AMD64_DIR%\
echo    - macOS Apple Silicon: %DARWIN_ARM64_DIR%\
echo.

REM ============================================
REM Windows 64-bit
REM ============================================
echo.
echo ════════════════════════════════════════════
echo Сборка для Windows 64-bit
echo ════════════════════════════════════════════

echo Сборка CLI версии для Windows 64-bit...
go build %BUILD_TAGS% -ldflags="-s -w" -o "%WINDOWS_DIR%\network-scanner.exe" ./cmd/network-scanner
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Ошибка сборки CLI версии для Windows
    exit /b 1
)
echo ✅ Собрано: %WINDOWS_DIR%\network-scanner.exe

echo Сборка GUI версии для Windows 64-bit...
go build %BUILD_TAGS% -ldflags="-s -w -H windowsgui" -o "%WINDOWS_DIR%\network-scanner-gui.exe" ./cmd/gui
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Ошибка сборки GUI версии для Windows
    exit /b 1
)
echo ✅ Собрано: %WINDOWS_DIR%\network-scanner-gui.exe

REM ============================================
REM Linux 64-bit
REM ============================================
echo.
echo ════════════════════════════════════════════
echo Сборка для Linux 64-bit
echo ════════════════════════════════════════════

set GOOS=linux
set GOARCH=amd64

echo Сборка CLI версии для Linux 64-bit...
go build %BUILD_TAGS% -ldflags="-s -w" -o "%LINUX_DIR%\network-scanner" ./cmd/network-scanner
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Ошибка сборки CLI версии для Linux
    set GOOS=
    set GOARCH=
    exit /b 1
)
echo ✅ Собрано: %LINUX_DIR%\network-scanner

echo Сборка GUI версии для Linux 64-bit...
go build %BUILD_TAGS% -ldflags="-s -w" -o "%LINUX_DIR%\network-scanner-gui" ./cmd/gui
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Ошибка сборки GUI версии для Linux
    set GOOS=
    set GOARCH=
    exit /b 1
)
echo ✅ Собрано: %LINUX_DIR%\network-scanner-gui

REM ============================================
REM macOS Intel (amd64)
REM ============================================
echo.
echo ════════════════════════════════════════════
echo Сборка для macOS Intel (amd64)
echo ════════════════════════════════════════════

set GOOS=darwin
set GOARCH=amd64

echo Сборка CLI версии для macOS Intel...
go build %BUILD_TAGS% -ldflags="-s -w" -o "%DARWIN_AMD64_DIR%\network-scanner" ./cmd/network-scanner
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Ошибка сборки CLI версии для macOS Intel
    set GOOS=
    set GOARCH=
    exit /b 1
)
echo ✅ Собрано: %DARWIN_AMD64_DIR%\network-scanner

echo Сборка GUI версии для macOS Intel...
go build %BUILD_TAGS% -ldflags="-s -w" -o "%DARWIN_AMD64_DIR%\network-scanner-gui" ./cmd/gui
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Ошибка сборки GUI версии для macOS Intel
    set GOOS=
    set GOARCH=
    exit /b 1
)
echo ✅ Собрано: %DARWIN_AMD64_DIR%\network-scanner-gui

REM ============================================
REM macOS Apple Silicon (arm64)
REM ============================================
echo.
echo ════════════════════════════════════════════
echo Сборка для macOS Apple Silicon (arm64)
echo ════════════════════════════════════════════

set GOOS=darwin
set GOARCH=arm64

echo Сборка CLI версии для macOS Apple Silicon...
go build %BUILD_TAGS% -ldflags="-s -w" -o "%DARWIN_ARM64_DIR%\network-scanner" ./cmd/network-scanner
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Ошибка сборки CLI версии для macOS Apple Silicon
    set GOOS=
    set GOARCH=
    exit /b 1
)
echo ✅ Собрано: %DARWIN_ARM64_DIR%\network-scanner

echo Сборка GUI версии для macOS Apple Silicon...
go build %BUILD_TAGS% -ldflags="-s -w" -o "%DARWIN_ARM64_DIR%\network-scanner-gui" ./cmd/gui
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Ошибка сборки GUI версии для macOS Apple Silicon
    set GOOS=
    set GOARCH=
    exit /b 1
)
echo ✅ Собрано: %DARWIN_ARM64_DIR%\network-scanner-gui

REM Сбрасываем переменные окружения
set GOOS=
set GOARCH=

REM ============================================
REM Копирование документации
REM ============================================
echo.
echo ════════════════════════════════════════════
echo Копирование документации
echo ════════════════════════════════════════════

REM Копируем инструкцию по эксплуатации в основную директорию и каждую поддиректорию
if exist "Инструкция по эксплуатации.md" (
    copy "Инструкция по эксплуатации.md" "%RELEASE_DIR%\" >nul
    copy "Инструкция по эксплуатации.md" "%WINDOWS_DIR%\" >nul
    copy "Инструкция по эксплуатации.md" "%LINUX_DIR%\" >nul
    copy "Инструкция по эксплуатации.md" "%DARWIN_AMD64_DIR%\" >nul
    copy "Инструкция по эксплуатации.md" "%DARWIN_ARM64_DIR%\" >nul
    echo ✅ Инструкция по эксплуатации скопирована во все директории
) else (
    echo ⚠️  Файл 'Инструкция по эксплуатации.md' не найден в корне проекта
)

REM Создаем README для каждой платформы
echo.
echo Создание README файлов для каждой платформы...

REM Windows README
(
echo # Network Scanner для Windows
echo.
echo ## Установка
echo.
echo Просто скачайте и запустите нужный файл:
echo.
echo - **network-scanner.exe** - Консольная версия
echo - **network-scanner-gui.exe** - Версия с графическим интерфейсом
echo.
echo ## Использование
echo.
echo ### GUI версия
echo Дважды кликните на `network-scanner-gui.exe` для запуска графического интерфейса.
echo.
echo ### CLI версия
echo Запустите `network-scanner.exe` из командной строки:
echo.
echo ```bash
echo network-scanner.exe
echo network-scanner.exe --network 192.168.1.0/24
echo network-scanner.exe --ports 80,443,8080 --threads 200
echo ```
echo.
echo Подробная документация находится в файле "Инструкция по эксплуатации.md"
) > "%WINDOWS_DIR%\README.md"

REM Linux README
(
echo # Network Scanner для Linux
echo.
echo ## Установка
echo.
echo 1. Скачайте файл для вашей архитектуры
echo 2. Сделайте файл исполняемым:
echo.
echo ```bash
echo chmod +x network-scanner
echo chmod +x network-scanner-gui
echo ```
echo.
echo ## Использование
echo.
echo ### GUI версия
echo ```bash
echo ./network-scanner-gui
echo ```
echo.
echo ### CLI версия
echo ```bash
echo ./network-scanner
echo ./network-scanner --network 192.168.1.0/24
echo ./network-scanner --ports 80,443,8080 --threads 200
echo ```
echo.
echo Подробная документация находится в файле "Инструкция по эксплуатации.md"
) > "%LINUX_DIR%\README.md"

REM macOS README
(
echo # Network Scanner для macOS
echo.
echo ## Установка
echo.
echo 1. Скачайте файл для вашей архитектуры:
echo    - **darwin-amd64** - для Intel Mac
echo    - **darwin-arm64** - для Apple Silicon (M1, M2, M3)
echo.
echo 2. Сделайте файл исполняемым:
echo.
echo ```bash
echo chmod +x network-scanner
echo chmod +x network-scanner-gui
echo ```
echo.
echo 3. Если macOS блокирует запуск, разрешите в настройках:
echo    Системные настройки ^> Безопасность и конфиденциальность
echo.
echo ## Использование
echo.
echo ### GUI версия
echo ```bash
echo ./network-scanner-gui
echo ```
echo.
echo ### CLI версия
echo ```bash
echo ./network-scanner
echo ./network-scanner --network 192.168.1.0/24
echo ./network-scanner --ports 80,443,8080 --threads 200
echo ```
echo.
echo Подробная документация находится в файле "Инструкция по эксплуатации.md"
) > "%DARWIN_AMD64_DIR%\README.md"
copy "%DARWIN_AMD64_DIR%\README.md" "%DARWIN_ARM64_DIR%\README.md" >nul

echo ✅ README файлы созданы для всех платформ

REM ============================================
REM Итоговая информация
REM ============================================
echo.
echo ════════════════════════════════════════════
if /i "%BUILD_TYPE%"=="debug" (
    echo ✅ Тестовая сборка завершена!
    echo ℹ️  Логи будут записываться в файлы LOG-*.txt
) else (
    echo ✅ Релизная сборка завершена!
)
echo ════════════════════════════════════════════
echo.
echo 📦 Структура релиза:
echo    %RELEASE_DIR%\
echo    ├── windows\
echo    │   ├── network-scanner.exe
echo    │   ├── network-scanner-gui.exe
echo    │   ├── README.md
echo    │   └── Инструкция по эксплуатации.md
echo    ├── linux\
echo    │   ├── network-scanner
echo    │   ├── network-scanner-gui
echo    │   ├── README.md
echo    │   └── Инструкция по эксплуатации.md
echo    ├── darwin-amd64\
echo    │   ├── network-scanner
echo    │   ├── network-scanner-gui
echo    │   ├── README.md
echo    │   └── Инструкция по эксплуатации.md
echo    ├── darwin-arm64\
echo    │   ├── network-scanner
echo    │   ├── network-scanner-gui
echo    │   ├── README.md
echo    │   └── Инструкция по эксплуатации.md
echo    └── Инструкция по эксплуатации.md
echo.
echo Все бинарники готовы к распространению!
echo.
