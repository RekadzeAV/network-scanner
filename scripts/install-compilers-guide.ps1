# Инструкция по установке компиляторов для кросскомпиляции
# Network Scanner - Установка кросс-компиляторов

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Установка компиляторов для сборки" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Проверка текущего состояния
Write-Host "[Проверка] Текущее состояние компиляторов..." -ForegroundColor Yellow

$gcc = Get-Command gcc -ErrorAction SilentlyContinue
if ($gcc) {
    Write-Host "  ✅ GCC для Windows: установлен" -ForegroundColor Green
    gcc --version | Select-Object -First 1
} else {
    Write-Host "  ❌ GCC для Windows: не найден" -ForegroundColor Red
}

$linuxGcc = Get-Command x86_64-linux-gnu-gcc -ErrorAction SilentlyContinue
$muslGcc = Get-Command x86_64-linux-musl-gcc -ErrorAction SilentlyContinue

if ($linuxGcc -or $muslGcc) {
    Write-Host "  ✅ Кросс-компилятор для Linux: установлен" -ForegroundColor Green
} else {
    Write-Host "  ⚠️  Кросс-компилятор для Linux: не найден" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Варианты установки" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "ПРОБЛЕМА:" -ForegroundColor Red
Write-Host "  Установка кросс-компилятора для Linux на Windows сталкивается" -ForegroundColor Yellow
Write-Host "  с проблемой длинных путей и символьных ссылок в tar архивах." -ForegroundColor Yellow
Write-Host ""

Write-Host "РЕКОМЕНДУЕМЫЕ РЕШЕНИЯ:" -ForegroundColor Green
Write-Host ""

Write-Host "1. Использовать WSL (Windows Subsystem for Linux)" -ForegroundColor Cyan
Write-Host "   - Установить WSL: wsl --install" -ForegroundColor Gray
Write-Host "   - В WSL установить Go и собрать Linux версию" -ForegroundColor Gray
Write-Host "   - Это самый простой способ для Linux сборки" -ForegroundColor Gray
Write-Host ""

Write-Host "2. Использовать Docker (если установлен)" -ForegroundColor Cyan
Write-Host "   - Создать Dockerfile для сборки Linux версии" -ForegroundColor Gray
Write-Host "   - Запустить сборку в контейнере" -ForegroundColor Gray
Write-Host ""

Write-Host "3. Использовать CI/CD (GitHub Actions)" -ForegroundColor Cyan
Write-Host "   - Настроить автоматическую сборку для всех платформ" -ForegroundColor Gray
Write-Host "   - Linux версия будет собираться на Linux runner" -ForegroundColor Gray
Write-Host "   - macOS версия будет собираться на macOS runner" -ForegroundColor Gray
Write-Host ""

Write-Host "4. Ручная установка кросс-компилятора" -ForegroundColor Cyan
Write-Host "   - Скачать архив с https://musl.cc/" -ForegroundColor Gray
Write-Host "   - Распаковать через 7-Zip или WSL" -ForegroundColor Gray
Write-Host "   - Добавить bin директорию в PATH" -ForegroundColor Gray
Write-Host ""

Write-Host "5. Сборка только для Windows (текущее состояние)" -ForegroundColor Cyan
Write-Host "   - Windows версии уже собираются успешно" -ForegroundColor Gray
Write-Host "   - Linux и macOS версии можно собрать на соответствующих платформах" -ForegroundColor Gray
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Текущий статус" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "✅ Windows сборка: ГОТОВА" -ForegroundColor Green
Write-Host "   - GCC установлен" -ForegroundColor Gray
Write-Host "   - CGO включен" -ForegroundColor Gray
Write-Host "   - Можно собирать Windows версии" -ForegroundColor Gray
Write-Host ""

Write-Host "⚠️  Linux сборка: ТРЕБУЕТСЯ КРОСС-КОМПИЛЯТОР" -ForegroundColor Yellow
Write-Host "   - Рекомендуется использовать WSL или CI/CD" -ForegroundColor Gray
Write-Host ""

Write-Host "❌ macOS сборка: НЕВОЗМОЖНА НА WINDOWS" -ForegroundColor Red
Write-Host "   - Требуется macOS SDK (доступен только на macOS)" -ForegroundColor Gray
Write-Host "   - Используйте macOS машину или CI/CD" -ForegroundColor Gray
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Быстрая установка WSL (рекомендуется)" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$wslInstalled = wsl --list --quiet 2>&1
if ($LASTEXITCODE -eq 0 -and $wslInstalled) {
    Write-Host "✅ WSL уже установлен!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Для сборки Linux версии в WSL:" -ForegroundColor Yellow
    Write-Host "  wsl" -ForegroundColor Gray
    Write-Host "  cd /mnt/d/Разработка` через` ИИ/network-scanner" -ForegroundColor Gray
    Write-Host "  go build -o network-scanner-linux ./cmd/network-scanner" -ForegroundColor Gray
    Write-Host "  go build -o network-scanner-gui-linux ./cmd/gui" -ForegroundColor Gray
} else {
    Write-Host "WSL не установлен. Для установки:" -ForegroundColor Yellow
    Write-Host "  wsl --install" -ForegroundColor Gray
    Write-Host ""
    Write-Host "⚠️  Требуются права администратора" -ForegroundColor Yellow
}

Write-Host ""
