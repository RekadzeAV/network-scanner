@echo off
REM Скрипт для сборки сканера сети для Windows

echo Сборка Network Scanner...

REM Создаем директорию для бинарников
if not exist dist mkdir dist

REM Текущая платформа (Windows)
echo Сборка для Windows 64-bit...
go build -o dist\network-scanner.exe ./cmd/network-scanner

REM Linux 64-bit
echo Сборка для Linux 64-bit...
set GOOS=linux
set GOARCH=amd64
go build -o dist\network-scanner-linux-amd64 ./cmd/network-scanner

REM macOS Intel
echo Сборка для macOS Intel...
set GOOS=darwin
set GOARCH=amd64
go build -o dist\network-scanner-darwin-amd64 ./cmd/network-scanner

REM macOS Apple Silicon
echo Сборка для macOS Apple Silicon...
set GOOS=darwin
set GOARCH=arm64
go build -o dist\network-scanner-darwin-arm64 ./cmd/network-scanner

echo Сборка завершена! Бинарники находятся в директории dist\

