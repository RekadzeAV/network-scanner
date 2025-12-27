@echo off
REM Скрипт для сборки сканера сети для Windows

echo Сборка Network Scanner...

REM Создаем директорию для бинарников
set RELEASE_DIR=release
if not exist "%RELEASE_DIR%" mkdir "%RELEASE_DIR%"
echo 📦 Бинарники будут сохранены в: %RELEASE_DIR%\
echo.

REM Текущая платформа (Windows)
echo Сборка для Windows 64-bit...
go build -ldflags="-s -w" -o "%RELEASE_DIR%\network-scanner-gui.exe" ./cmd/gui

REM Linux 64-bit
echo Сборка для Linux 64-bit...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o "%RELEASE_DIR%\network-scanner-gui-linux-amd64" ./cmd/gui

REM macOS Intel
echo Сборка для macOS Intel...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags="-s -w" -o "%RELEASE_DIR%\network-scanner-gui-darwin-amd64" ./cmd/gui

REM macOS Apple Silicon
echo Сборка для macOS Apple Silicon...
set GOOS=darwin
set GOARCH=arm64
go build -ldflags="-s -w" -o "%RELEASE_DIR%\network-scanner-gui-darwin-arm64" ./cmd/gui

echo.
echo ✅ Сборка завершена! Бинарники находятся в директории %RELEASE_DIR%\

