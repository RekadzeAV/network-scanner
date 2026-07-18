# Скрипт для сборки Network Scanner под Android
#
# Требования:
# - Android SDK (ANDROID_HOME)
# - JDK 17+
# - gomobile установлен: go install golang.org/x/mobile/cmd/gomobile@latest

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectDir = Split-Path -Parent $ScriptDir
$OutputDir = Join-Path $ScriptDir "..\release-build"

Write-Host "📱 Сборка Network Scanner для Android..." -ForegroundColor Cyan

# Проверяем наличие ANDROID_HOME
if (-not (Test-Path env:ANDROID_HOME) -and -not (Test-Path env:ANDROID_SDK_ROOT)) {
    Write-Host "❌ Ошибка: ANDROID_HOME не установлен" -ForegroundColor Red
    Write-Host "Установите Android SDK и задайте переменную окружения" -ForegroundColor Yellow
    exit 1
}

# Создаем директорию вывода
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

# Очищаем предыдущую сборку
Remove-Item (Join-Path $OutputDir "network-scanner-android.*") -ErrorAction SilentlyContinue

# Сборка
Set-Location $ProjectDir

$GOMOBILE = Join-Path $env:USERPROFILE "go\bin\gomobile.exe"
if (-not (Test-Path $GOMOBILE)) {
    Write-Host "❌ Gomobile не найден. Установите: go install golang.org/x/mobile/cmd/gomobile@latest" -ForegroundColor Red
    exit 1
}

Write-Host "⏳ Запуск gomobile bind..." -ForegroundColor Yellow

& $GOMOBILE bind -target=android -o="NetworkScanner" `
    -javapkg=network.scanner `
    -androidapi=21 `
    .\...

# Переименовываем APK
if (Test-Path "NetworkScanner.aar") {
    Move-Item "NetworkScanner.aar" (Join-Path $OutputDir "network-scanner-android.aar") -Force
    Write-Host "✅ Сборка завершена: $OutputDir\network-scanner-android.aar" -ForegroundColor Green
}

if (Test-Path "NetworkScanner-sources.jar") {
    Move-Item "NetworkScanner-sources.jar" (Join-Path $OutputDir "network-scanner-android-sources.jar") -Force
    Write-Host "✅ Источники: $OutputDir\network-scanner-android-sources.jar" -ForegroundColor Green
}

Write-Host "📦 Артефакты собраны в: $OutputDir" -ForegroundColor Green
