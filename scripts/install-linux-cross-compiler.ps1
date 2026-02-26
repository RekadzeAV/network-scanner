# Скрипт для установки кросс-компилятора для Linux на Windows
# Этот скрипт помогает установить x86_64-linux-musl-gcc для кросскомпиляции

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Установка кросс-компилятора для Linux" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$installDir = "C:\cross-compilers"
$archiveUrl = "https://musl.cc/x86_64-linux-musl-cross.tgz"
$archivePath = "$env:TEMP\x86_64-linux-musl-cross.tgz"

Write-Host "[1/4] Скачивание кросс-компилятора..." -ForegroundColor Yellow
try {
    $ProgressPreference = 'SilentlyContinue'
    Invoke-WebRequest -Uri $archiveUrl -OutFile $archivePath
    Write-Host "  ✅ Архив скачан: $archivePath" -ForegroundColor Green
} catch {
    Write-Host "  ❌ Ошибка при скачивании: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "[2/4] Создание директории установки..." -ForegroundColor Yellow
try {
    if (Test-Path $installDir) {
        Write-Host "  ⚠️  Директория уже существует: $installDir" -ForegroundColor Yellow
        $response = Read-Host "  Удалить и переустановить? (y/N)"
        if ($response -eq 'y' -or $response -eq 'Y') {
            Remove-Item -Recurse -Force $installDir
        } else {
            Write-Host "  Прервано пользователем" -ForegroundColor Yellow
            exit 0
        }
    }
    New-Item -ItemType Directory -Force -Path $installDir | Out-Null
    Write-Host "  ✅ Директория создана: $installDir" -ForegroundColor Green
} catch {
    Write-Host "  ❌ Ошибка при создании директории: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "[3/4] Распаковка архива..." -ForegroundColor Yellow
Write-Host "  ⚠️  ВНИМАНИЕ: Распаковка может занять некоторое время" -ForegroundColor Yellow
Write-Host "  ⚠️  Если возникнут ошибки с длинными путями, используйте WSL или 7-Zip" -ForegroundColor Yellow
Write-Host ""

# Попытка распаковки через tar
try {
    # Переходим в директорию установки
    Push-Location $installDir
    
    # Пробуем распаковать
    tar -xzf $archivePath 2>&1 | Out-Null
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  ✅ Архив успешно распакован" -ForegroundColor Green
    } else {
        Write-Host "  ⚠️  Ошибка при распаковке через tar" -ForegroundColor Yellow
        Write-Host "  Попробуйте распаковать вручную:" -ForegroundColor Yellow
        Write-Host "    1. Используйте 7-Zip для распаковки $archivePath" -ForegroundColor Gray
        Write-Host "    2. Или используйте WSL: wsl tar -xzf $archivePath -C /mnt/c/cross-compilers" -ForegroundColor Gray
        Write-Host "    3. Распакуйте в: $installDir" -ForegroundColor Gray
        Pop-Location
        exit 1
    }
    
    Pop-Location
} catch {
    Write-Host "  ❌ Ошибка: $_" -ForegroundColor Red
    Pop-Location
    exit 1
}

Write-Host ""
Write-Host "[4/4] Настройка PATH..." -ForegroundColor Yellow

# Определяем путь к бинарникам
$binPath = Join-Path $installDir "x86_64-linux-musl-cross\bin"

if (Test-Path $binPath) {
    # Проверяем, есть ли уже в PATH
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    
    if ($currentPath -notlike "*$binPath*") {
        # Добавляем в PATH пользователя
        $newPath = $currentPath + ";" + $binPath
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        Write-Host "  ✅ Путь добавлен в PATH пользователя" -ForegroundColor Green
        Write-Host "  ⚠️  Перезапустите терминал для применения изменений" -ForegroundColor Yellow
    } else {
        Write-Host "  ✅ Путь уже в PATH" -ForegroundColor Green
    }
    
    # Добавляем в текущую сессию
    $env:Path += ";$binPath"
    Write-Host "  ✅ Путь добавлен в текущую сессию" -ForegroundColor Green
} else {
    Write-Host "  ⚠️  Директория bin не найдена: $binPath" -ForegroundColor Yellow
    Write-Host "  Проверьте структуру распакованных файлов" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Проверка установки..." -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# Проверяем наличие компилятора
$gccPath = Join-Path $binPath "x86_64-linux-musl-gcc"
if (Test-Path $gccPath) {
    Write-Host ""
    Write-Host "✅ Кросс-компилятор установлен успешно!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Проверка версии:" -ForegroundColor Cyan
    & $gccPath --version 2>&1 | Select-Object -First 3
    
    Write-Host ""
    Write-Host "Использование для Go кросскомпиляции:" -ForegroundColor Cyan
    Write-Host '  $env:GOOS = "linux"' -ForegroundColor Gray
    Write-Host '  $env:GOARCH = "amd64"' -ForegroundColor Gray
    Write-Host '  $env:CC = "x86_64-linux-musl-gcc"' -ForegroundColor Gray
    Write-Host '  $env:CGO_ENABLED = "1"' -ForegroundColor Gray
    Write-Host '  go build -o app-linux ./cmd/gui' -ForegroundColor Gray
} else {
    Write-Host ""
    Write-Host "⚠️  Компилятор не найден по ожидаемому пути" -ForegroundColor Yellow
    Write-Host "  Ожидаемый путь: $gccPath" -ForegroundColor Gray
    Write-Host "  Проверьте структуру распакованных файлов в: $installDir" -ForegroundColor Gray
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan


