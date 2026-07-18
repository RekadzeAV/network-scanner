# save_benchmarks.ps1 — Запуск бенчмарков и сохранение результатов в artifacts
# Используется в CI для контроля регрессий производительности

$ErrorActionPreference = "Stop"

$ArtifactDir = "release-build\benchmarks"
New-Item -ItemType Directory -Force -Path $ArtifactDir | Out-Null

Write-Host "=== Запуск бенчмарков ===" -ForegroundColor Cyan

Write-Host "`n--- Scanner benchmarks ---" -ForegroundColor Yellow
go test -bench=. -benchmem -run=^$ -count=3 .\internal\scanner\ > "$ArtifactDir\scanner.txt" 2>&1

Write-Host "--- Network benchmarks ---" -ForegroundColor Yellow
go test -bench=. -benchmem -run=^$ -count=3 .\internal\network\ > "$ArtifactDir\network.txt" 2>&1

Write-Host "--- Topology benchmarks ---" -ForegroundColor Yellow
go test -bench=. -benchmem -run=^$ -count=3 .\internal\topology\ > "$ArtifactDir\topology.txt" 2>&1

Write-Host "`n=== Результаты сохранены в $ArtifactDir ===" -ForegroundColor Green
Get-ChildItem $ArtifactDir | Format-Table Name, Length

Write-Host "`n=== Краткая сводка ===" -ForegroundColor Green
Get-ChildItem "$ArtifactDir\*.txt" | ForEach-Object {
    Write-Host "`n--- $($_.Name) ---"
    Select-String -Path $_.FullName -Pattern "^Benchmark|^PASS|^ok" | Select-Object -First 20
}
