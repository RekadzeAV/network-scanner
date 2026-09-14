# =============================================================================
# Скрипт для сборки Docker-образа Network Scanner CLI (PowerShell)
# =============================================================================
# Использование:
#   .\scripts\build-docker.ps1              # Сборка с дефолтными параметрами
#   .\scripts\build-docker.ps1 2.3.0        # Сборка с указанной версией
#   .\scripts\build-docker.ps1 latest sha   # Сборка с версией и commit hash
# =============================================================================

param(
    [string]$Version = "dev",
    [string]$GitCommit = ""
)

$ErrorActionPreference = "Stop"

# Параметры
$BuildTime = Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ"
if (-not $GitCommit -or $GitCommit -eq "") {
    $GitCommit = git rev-parse --short HEAD 2>$null
    if (-not $GitCommit) { $GitCommit = "unknown" }
}

$ImageName = "network-scanner"
$ImageTag = $Version

Write-Host "=================================================================" -ForegroundColor Green
Write-Host "  Network Scanner CLI - Docker Build" -ForegroundColor Green
Write-Host "=================================================================" -ForegroundColor Green
Write-Host ""
Write-Host "  Версия:      $Version" -ForegroundColor Yellow
Write-Host "  Время сборки: $BuildTime" -ForegroundColor Yellow
Write-Host "  Commit:      $GitCommit" -ForegroundColor Yellow
Write-Host "  Образ:       $ImageName:$ImageTag" -ForegroundColor Yellow
Write-Host ""

# Проверка Docker
try {
    docker version > $null 2>&1
} catch {
    Write-Host "Ошибка: Docker не установлен. Установите Docker и попробуйте снова." -ForegroundColor Red
    exit 1
}

Write-Host ">>> Сборка Docker-образа..." -ForegroundColor Green
docker build `
    --build-arg Version="$Version" `
    --build-arg BuildTime="$BuildTime" `
    --build-arg GitCommit="$GitCommit" `
    -t "$ImageName:$ImageTag" `
    -f Dockerfile `
    .

Write-Host ""
Write-Host ">>> Проверка образа..." -ForegroundColor Green
docker image inspect "$ImageName:$ImageTag" > $null 2>&1
if ($?) {
    Write-Host "✅ Образ успешно собран!" -ForegroundColor Green
    
    Write-Host ""
    Write-Host ">>> Информация об образе:" -ForegroundColor Green
    docker images "$ImageName" --format "table {{.Repository}}`t{{.Tag}}`t{{.ID}}`t{{.Size}}`t{{.CreatedAt}}"
} else {
    Write-Host "❌ Ошибка проверки образа" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "=================================================================" -ForegroundColor Green
Write-Host "  Готово!" -ForegroundColor Green
Write-Host "=================================================================" -ForegroundColor Green
Write-Host ""
Write-Host "  Запуск CLI:" -ForegroundColor White
Write-Host "    docker run --rm $ImageName:$ImageTag --help"
Write-Host ""
Write-Host "  Запуск сканирования:" -ForegroundColor White
Write-Host "    docker run --rm --network host --cap-add NET_ADMIN --cap-add NET_RAW $ImageName:$ImageTag scan --cidr 192.168.1.0/24"
Write-Host ""
