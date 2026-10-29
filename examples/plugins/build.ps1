<#
.SYNOPSIS
  Сборка примеров плагинов Network Scanner в shared library (.dll на Windows).

.DESCRIPTION
  Каждый плагин собирается через `go build -buildmode=plugin` в подкаталог
  dist\<name>\. Требуются: включённый CGO и gcc (MinGW-w64) на Windows.
  Результат копируется рядом с plugin.json — так загрузчик (L2.2) находит плагин
  по NETWORK_SCANNER_PLUGINS_DIR.

.EXAMPLE
  .\build.ps1                 # собрать все плагины
  .\build.ps1 -Only onlyup    # собрать один плагин
#>
param(
    [string]$Only = ""
)

$ErrorActionPreference = "Stop"

# Каталог скрипта и корень модуля (примеры лежат в examples/plugins).
$PluginRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot = Resolve-Path (Join-Path $PluginRoot "..\..\..")

# Список плагинов: имя каталога -> имя entry-библиотеки из plugin.json.
$Plugins = @(
    @{ Dir = "onlyup";        Entry = "onlyup.dll" },
    @{ Dir = "json-exporter"; Entry = "jsonexporter.dll" },
    @{ Dir = "alert-report";  Entry = "alertreport.dll" }
)

if (-not (Get-Command gcc -ErrorAction SilentlyContinue)) {
    Write-Error "gcc не найден. Установите MinGW-w64 (w64devkit) и добавьте в PATH.`nБез CGO сборка -buildmode=plugin невозможна."
}

Push-Location $RepoRoot
try {
    foreach ($p in $Plugins) {
        if ($Only -and $p.Dir -ne $Only) { continue }

        $src = Join-Path $PluginRoot $p.Dir
        $outDir = Join-Path $src "dist"
        New-Item -ItemType Directory -Force -Path $outDir | Out-Null

        Write-Host "==> Сборка $($p.Dir)"
        $env:CGO_ENABLED = "1"
        # -buildmode=plugin требует package main + экспортируемый символ New.
        & go build -buildmode=plugin -o (Join-Path $outDir $p.Entry) $src
        if ($LASTEXITCODE -ne 0) {
            Write-Error "Сборка $($p.Dir) завершилась с кодом $LASTEXITCODE"
        }

        # Копируем библиотеку и манифест рядом — готовая плагин-директория.
        Copy-Item (Join-Path $src "plugin.json") $outDir -Force
        Write-Host "    готово: $outDir"
    }
}
finally {
    Pop-Location
}

Write-Host ""
Write-Host "Укажите каталог с dist\plugin.json в NETWORK_SCANNER_PLUGINS_DIR"
Write-Host "и запустите: network-scanner plugins list"
