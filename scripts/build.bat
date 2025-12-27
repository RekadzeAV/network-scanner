@echo off
REM Скрипт для сборки сканера сети для Windows

echo Сборка Network Scanner...

REM Создаем директорию для бинарников с датой сборки
for /f "tokens=2-4 delims=/ " %%a in ('date /t') do (set mydate=%%c-%%a-%%b)
for /f "tokens=1-2 delims=/ " %%a in ("%mydate%") do (
    set BUILD_DATE=%%a
)
if "%BUILD_DATE:~4,1%"=="/" set BUILD_DATE=%BUILD_DATE:~0,4%-%BUILD_DATE:~5,2%-%BUILD_DATE:~8,2%

REM Альтернативный способ получения даты в формате YYYY-MM-DD
for /f "tokens=2 delims==" %%I in ('wmic os get localdatetime /value') do set datetime=%%I
set BUILD_DATE=%datetime:~0,4%-%datetime:~4,2%-%datetime:~6,2%

set RELEASE_DIR=Release\%BUILD_DATE%
if not exist "%RELEASE_DIR%" mkdir "%RELEASE_DIR%"
echo 📦 Бинарники будут сохранены в: %RELEASE_DIR%\
echo.

REM Текущая платформа (Windows)
echo Сборка для Windows 64-bit...
go build -ldflags="-s -w" -o "%RELEASE_DIR%\network-scanner.exe" ./cmd/network-scanner

REM Linux 64-bit
echo Сборка для Linux 64-bit...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o "%RELEASE_DIR%\network-scanner-linux-amd64" ./cmd/network-scanner

REM macOS Intel
echo Сборка для macOS Intel...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags="-s -w" -o "%RELEASE_DIR%\network-scanner-darwin-amd64" ./cmd/network-scanner

REM macOS Apple Silicon
echo Сборка для macOS Apple Silicon...
set GOOS=darwin
set GOARCH=arm64
go build -ldflags="-s -w" -o "%RELEASE_DIR%\network-scanner-darwin-arm64" ./cmd/network-scanner

echo.
echo ✅ Сборка завершена! Бинарники находятся в директории %RELEASE_DIR%\

