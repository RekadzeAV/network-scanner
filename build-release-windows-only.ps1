# Скрипт для сборки релиза Network Scanner (только Windows)
$ErrorActionPreference = "Stop"

Write-Host "=========================================="
Write-Host "Сборка Network Scanner - Релиз (Windows)"
Write-Host "=========================================="
Write-Host ""

# Создаем директорию для бинарников с датой сборки и номером
$buildDate = Get-Date -Format "yyyy-MM-dd"
$buildNum = 1
$releaseDir = "release\$buildDate-$buildNum"

# Находим следующий доступный номер сборки
while (Test-Path $releaseDir) {
    $buildNum++
    $releaseDir = "release\$buildDate-$buildNum"
}

# Создаем директории
New-Item -ItemType Directory -Path $releaseDir -Force | Out-Null
$windowsDir = "$releaseDir\windows"
New-Item -ItemType Directory -Path $windowsDir -Force | Out-Null

Write-Host "Бинарники будут сохранены в: $releaseDir\ (сборка #$buildNum)"
Write-Host ""

# Установка зависимостей
Write-Host "Установка зависимостей..."
go mod download
go mod tidy
Write-Host "Зависимости установлены"
Write-Host ""

# ===========================================
# Windows 64-bit
# ===========================================
Write-Host "==========================================="
Write-Host "Сборка для Windows 64-bit"
Write-Host "==========================================="
Write-Host ""

Write-Host "Сборка CLI версии для Windows 64-bit..."
go build -ldflags="-s -w" -o "$windowsDir\network-scanner.exe" ./cmd/network-scanner
if ($LASTEXITCODE -ne 0) {
    Write-Host "Ошибка сборки CLI версии для Windows"
    exit 1
}
Write-Host "Собрано: $windowsDir\network-scanner.exe"

Write-Host "Сборка GUI версии для Windows 64-bit..."
go build -ldflags="-s -w -H windowsgui" -o "$windowsDir\network-scanner-gui.exe" ./cmd/gui
if ($LASTEXITCODE -ne 0) {
    Write-Host "Ошибка сборки GUI версии для Windows"
    exit 1
}
Write-Host "Собрано: $windowsDir\network-scanner-gui.exe"

# ===========================================
# Копирование документации
# ===========================================
Write-Host ""
Write-Host "==========================================="
Write-Host "Копирование документации"
Write-Host "==========================================="
Write-Host ""

if (Test-Path "Инструкция по эксплуатации.md") {
    Copy-Item "Инструкция по эксплуатации.md" "$releaseDir\"
    Copy-Item "Инструкция по эксплуатации.md" "$windowsDir\"
    Write-Host "Инструкция по эксплуатации скопирована"
} else {
    Write-Host "Файл 'Инструкция по эксплуатации.md' не найден в корне проекта"
}

Write-Host ""
Write-Host "=========================================="
Write-Host "Релизная сборка завершена!"
Write-Host "=========================================="
Write-Host ""
Write-Host "Структура релиза:"
Write-Host "   $releaseDir\"
Write-Host "   ├── windows\"
Write-Host "   │   ├── network-scanner.exe"
Write-Host "   │   ├── network-scanner-gui.exe"
Write-Host "   │   └── Инструкция по эксплуатации.md"
Write-Host "   └── Инструкция по эксплуатации.md"
Write-Host ""
Write-Host "Все бинарники готовы к распространению!"
Write-Host ""
