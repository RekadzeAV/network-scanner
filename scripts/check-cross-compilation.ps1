# Скрипт проверки готовности системы к кроссплатформенной сборке
# Проверяет наличие всех необходимых компонентов для сборки на Windows

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Проверка готовности к кроссплатформенной сборке" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$allChecksPassed = $true

# Проверка Go
Write-Host "[1/6] Проверка Go..." -ForegroundColor Yellow
try {
    $goVersion = go version 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  ✅ Go установлен" -ForegroundColor Green
        Write-Host "     $($goVersion -split "`n" | Select-Object -First 1)" -ForegroundColor Gray
        
        # Проверка версии
        if ($goVersion -match "go(\d+)\.(\d+)") {
            $major = [int]$matches[1]
            $minor = [int]$matches[2]
            if ($major -gt 1 -or ($major -eq 1 -and $minor -ge 21)) {
                Write-Host "     Версия соответствует требованиям (>= 1.21)" -ForegroundColor Green
            } else {
                Write-Host "     ⚠️  Версия ниже требуемой (нужна >= 1.21)" -ForegroundColor Yellow
                $allChecksPassed = $false
            }
        }
    } else {
        throw "Go не найден"
    }
} catch {
    Write-Host "  ❌ Go не установлен или не найден в PATH" -ForegroundColor Red
    Write-Host "     Установите Go с https://go.dev/dl/" -ForegroundColor Yellow
    $allChecksPassed = $false
}

Write-Host ""

# Проверка GCC для Windows
Write-Host "[2/6] Проверка GCC для Windows..." -ForegroundColor Yellow
try {
    $gccVersion = gcc --version 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  ✅ GCC установлен" -ForegroundColor Green
        $firstLine = ($gccVersion -split "`n" | Select-Object -First 1)
        Write-Host "     $firstLine" -ForegroundColor Gray
        
        # Проверка типа GCC
        if ($gccVersion -match "mingw|MinGW|TDM|tdm|WinLibs|winlibs") {
            Write-Host "     Тип: MinGW-w64 (подходит для Windows)" -ForegroundColor Green
        } else {
            Write-Host "     ⚠️  Неизвестный тип GCC" -ForegroundColor Yellow
        }
    } else {
        throw "GCC не найден"
    }
} catch {
    Write-Host "  ❌ GCC не найден в PATH" -ForegroundColor Red
    Write-Host "     Установите MinGW-w64, TDM-GCC или WinLibs" -ForegroundColor Yellow
    Write-Host "     Варианты:" -ForegroundColor Yellow
    Write-Host "       - WinLibs: winget install BrechtSanders.WinLibs.POSIX.UCRT" -ForegroundColor Gray
    Write-Host "       - TDM-GCC: https://jmeubank.github.io/tdm-gcc/" -ForegroundColor Gray
    Write-Host "       - MSYS2: https://www.msys2.org/" -ForegroundColor Gray
    $allChecksPassed = $false
}

Write-Host ""

# Проверка CGO
Write-Host "[3/6] Проверка CGO..." -ForegroundColor Yellow
try {
    $cgoEnabled = go env CGO_ENABLED 2>&1
    if ($LASTEXITCODE -eq 0) {
        if ($cgoEnabled -eq "1") {
            Write-Host "  ✅ CGO включен" -ForegroundColor Green
            Write-Host "     CGO необходим для работы Fyne GUI framework" -ForegroundColor Gray
        } else {
            Write-Host "  ⚠️  CGO отключен" -ForegroundColor Yellow
            Write-Host "     Включите CGO: `$env:CGO_ENABLED = '1'" -ForegroundColor Yellow
            Write-Host "     Или: go env -w CGO_ENABLED=1" -ForegroundColor Yellow
            $allChecksPassed = $false
        }
    } else {
        throw "Не удалось проверить CGO"
    }
} catch {
    Write-Host "  ❌ Ошибка при проверке CGO" -ForegroundColor Red
    $allChecksPassed = $false
}

Write-Host ""

# Проверка кросс-компилятора для Linux
Write-Host "[4/6] Проверка кросс-компилятора для Linux..." -ForegroundColor Yellow
$linuxGcc = Get-Command x86_64-linux-gnu-gcc -ErrorAction SilentlyContinue
if ($linuxGcc) {
    try {
        $linuxGccVersion = & $linuxGcc.FullName --version 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "  ✅ x86_64-linux-gnu-gcc найден" -ForegroundColor Green
            $firstLine = ($linuxGccVersion -split "`n" | Select-Object -First 1)
            Write-Host "     $firstLine" -ForegroundColor Gray
            Write-Host "     Путь: $($linuxGcc.FullName)" -ForegroundColor Gray
        }
    } catch {
        Write-Host "  ⚠️  x86_64-linux-gnu-gcc найден, но не работает" -ForegroundColor Yellow
        $allChecksPassed = $false
    }
} else {
    Write-Host "  ⚠️  Кросс-компилятор для Linux не найден" -ForegroundColor Yellow
    Write-Host "     Для кросскомпиляции Linux на Windows требуется:" -ForegroundColor Yellow
    Write-Host "       x86_64-linux-gnu-gcc" -ForegroundColor Gray
    Write-Host "     Установка через MSYS2:" -ForegroundColor Yellow
    Write-Host "       pacman -S mingw-w64-x86_64-gcc-linux-gnu" -ForegroundColor Gray
    Write-Host "     Альтернатива: используйте CI/CD для сборки Linux версии" -ForegroundColor Yellow
    # Не считаем это критической ошибкой, так как можно использовать CI/CD
}

Write-Host ""

# Проверка musl-gcc (альтернатива для статической сборки Linux)
Write-Host "[5/6] Проверка musl-gcc (опционально, для статической сборки)..." -ForegroundColor Yellow
$muslGcc = Get-Command x86_64-linux-musl-gcc -ErrorAction SilentlyContinue
if ($muslGcc) {
    Write-Host "  ✅ x86_64-linux-musl-gcc найден" -ForegroundColor Green
    Write-Host "     Можно создавать статические Linux бинарники" -ForegroundColor Gray
} else {
    Write-Host "  ℹ️  musl-gcc не установлен (не обязательно)" -ForegroundColor Gray
    Write-Host "     Полезен для создания статических Linux бинарников" -ForegroundColor Gray
}

Write-Host ""

# Проверка инструментов для macOS
Write-Host "[6/6] Проверка инструментов для macOS..." -ForegroundColor Yellow
Write-Host "  ⚠️  Кросскомпиляция macOS на Windows невозможна" -ForegroundColor Yellow
Write-Host "     Причины:" -ForegroundColor Yellow
Write-Host "       - macOS SDK доступен только на macOS (лицензионное ограничение)" -ForegroundColor Gray
Write-Host "       - Даже при наличии SDK, кросс-компиляция крайне сложна" -ForegroundColor Gray
Write-Host "     Рекомендации:" -ForegroundColor Yellow
Write-Host "       - Используйте macOS машину для сборки" -ForegroundColor Gray
Write-Host "       - Используйте CI/CD (GitHub Actions с macos-latest runner)" -ForegroundColor Gray
Write-Host "       - Используйте виртуальную машину macOS (если есть лицензия)" -ForegroundColor Gray

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan

# Итоговый результат
if ($allChecksPassed) {
    Write-Host "✅ Основные компоненты готовы к сборке для Windows" -ForegroundColor Green
    Write-Host ""
    Write-Host "Для кроссплатформенной сборки:" -ForegroundColor Cyan
    Write-Host "  - Windows: ✅ Готово" -ForegroundColor Green
    Write-Host "  - Linux: ⚠️  Требуется кросс-компилятор (или используйте CI/CD)" -ForegroundColor Yellow
    Write-Host "  - macOS: ❌ Используйте macOS машину или CI/CD" -ForegroundColor Red
} else {
    Write-Host "❌ Некоторые обязательные компоненты отсутствуют" -ForegroundColor Red
    Write-Host "   Установите недостающие компоненты перед сборкой" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Подробная информация: docs/CROSS_COMPILATION_WINDOWS.md" -ForegroundColor Gray
Write-Host "========================================" -ForegroundColor Cyan

