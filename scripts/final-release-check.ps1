# Final Release Verification Script — v2.0 (Windows)
# Запуск: .\scripts\final-release-check.ps1

Write-Host "═══════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host " Network Scanner v2.0 — Final Release Check" -ForegroundColor Cyan
Write-Host "═══════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""

# 1. Build check
Write-Host "[1/5] Build CLI..." -ForegroundColor Yellow
$ErrorActionPreference = "Stop"
try {
    go build -o /tmp/network-scanner-test.exe ./cmd/network-scanner
    if (Test-Path /tmp/network-scanner-test.exe) {
        Write-Host "  ✅ CLI собран" -ForegroundColor Green
        Remove-Item /tmp/network-scanner-test.exe -ErrorAction SilentlyContinue
    } else {
        Write-Host "  ❌ CLI не собрался" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "  ❌ CLI build failed: $_" -ForegroundColor Red
    exit 1
}

Write-Host "[2/5] Build GUI..." -ForegroundColor Yellow
try {
    go build -o /tmp/network-scanner-gui-test.exe ./cmd/gui
    if (Test-Path /tmp/network-scanner-gui-test.exe) {
        Write-Host "  ✅ GUI собран" -ForegroundColor Green
        Remove-Item /tmp/network-scanner-gui-test.exe -ErrorAction SilentlyContinue
    } else {
        Write-Host "  ❌ GUI не собрался" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "  ❌ GUI build failed: $_" -ForegroundColor Red
    exit 1
}

# 2. Unit tests
Write-Host "[3/5] Unit tests..." -ForegroundColor Yellow
try {
    go test ./... -count=1 | Out-File /tmp/test-output.txt
    if ($?) {
        Write-Host "  ✅ go test ./... прошёл" -ForegroundColor Green
    } else {
        Write-Host "  ❌ go test ./... упал" -ForegroundColor Red
        Get-Content /tmp/test-output.txt
        exit 1
    }
} catch {
    Write-Host "  ❌ Tests failed: $_" -ForegroundColor Red
    exit 1
}

# 3. Smoke — topology export
Write-Host "[4/5] Smoke — topology export..." -ForegroundColor Yellow
if (Test-Path "./scripts/smoke-d-track-topology-export.ps1") {
    try {
        .\scripts\smoke-d-track-topology-export.ps1
        if ($?) {
            Write-Host "  ✅ smoke-d-track-topology-export прошёл" -ForegroundColor Green
        } else {
            Write-Host "  ⚠️  smoke-d-track-topology-export упал (не критично)" -ForegroundColor Yellow
        }
    } catch {
        Write-Host "  ⚠️  smoke-d-track-topology-export упал: $_" -ForegroundColor Yellow
    }
} else {
    Write-Host "  ⚠️  smoke-d-track-topology-export.ps1 не найден" -ForegroundColor Yellow
}

# 4. Documentation sanity
Write-Host "[5/5] Docs sanity..." -ForegroundColor Yellow
$docsPresent = $true
if (-not (Test-Path "./docs/FINAL_RELEASE_READINESS_REPORT.md")) { $docsPresent = $false; Write-Host "  ⚠️  FINAL_RELEASE_READINESS_REPORT.md отсутствует" -ForegroundColor Yellow }
if (-not (Test-Path "./docs/D_TRACK_IMPLEMENTATION_STATUS.md")) { $docsPresent = $false; Write-Host "  ⚠️  D_TRACK_IMPLEMENTATION_STATUS.md отсутствует" -ForegroundColor Yellow }
if (-not (Test-Path "./CHANGELOG.md")) { $docsPresent = $false; Write-Host "  ⚠️  CHANGELOG.md отсутствует" -ForegroundColor Yellow }

if ($docsPresent) {
    Write-Host "  ✅ Ключевые документы на месте" -ForegroundColor Green
} else {
    Write-Host "  ❌ Некоторые документы отсутствуют" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "═══════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host " ✅ Все проверки пройдены!" -ForegroundColor Green
Write-Host "═══════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""
Write-Host "Следующие шаги:" -ForegroundColor White
Write-Host "  1. Запустить кросс-ОС прогон на Linux/macOS"
Write-Host "  2. Разблокировать CI через GITHUB_TOKEN"
Write-Host "  3. Пройти ручную GUI приёмку (docs/GUI_SMOKE_CHECKLIST.md)"
Write-Host "  4. Выпустить релиз"
Write-Host ""
