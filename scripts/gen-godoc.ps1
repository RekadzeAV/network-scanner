# Docs generation: godoc HTML для внутренних пакетов (M3).
#
# Генерирует статический HTML-сайт godoc в docs/dev/godoc/.
# Требует: Go 1.25+ и `go install golang.org/x/tools/cmd/godoc@latest`
# (или использует `go doc` для текстового дампа, если godoc недоступен).
#
# Usage (PowerShell):
#   .\scripts\gen-godoc.ps1

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$outDir = Join-Path $repoRoot "docs\dev\godoc"

Write-Host "==> Generating godoc into $outDir"

if (-not (Test-Path $outDir)) {
    New-Item -ItemType Directory -Path $outDir -Force | Out-Null
}

# Пытаемся использовать классический godoc, если он установлен.
$godoc = Get-Command godoc -ErrorAction SilentlyContinue
if ($godoc) {
    Push-Location $repoRoot
    try {
        & godoc -http=:0 -goroot=. -path="$repoRoot" -write_index_to="$outDir" 2>&1 | Out-Null
        Write-Host "==> godoc index written to $outDir"
    } catch {
        Write-Warning "godoc exited with error: $_"
    } finally {
        Pop-Location
    }
    exit 0
}

Write-Warning "godoc not installed; falling back to 'go doc' text dump."

# Собираем директории пакетов из internal/ и cmd/.
# Пакет считается валидным, если содержит хотя бы один .go-файл, НЕ являющийся _test.go.
function Get-PackageDirs {
    param([string]$BaseDir, [string]$Prefix)
    $result = @()
    if (-not (Test-Path $BaseDir)) { return $result }
    $dirs = Get-ChildItem -Path $BaseDir -Directory -Recurse -ErrorAction SilentlyContinue
    foreach ($d in $dirs) {
        $goFiles = Get-ChildItem -Path $d.FullName -Filter "*.go" -File -ErrorAction SilentlyContinue |
            Where-Object { $_.Name -notlike "*_test.go" }
        if ($goFiles.Count -gt 0) {
            $rel = $d.FullName.Substring($BaseDir.Length).TrimStart('\','/')
            $rel = $rel.Replace('\', '/')
            $importPath = "network-scanner/$Prefix/$rel"
            $result += [PSCustomObject]@{ Dir = $d; ImportPath = $importPath }
        }
    }
    return $result
}

$allPackages = @()
$allPackages += Get-PackageDirs -BaseDir (Join-Path $repoRoot "internal") -Prefix "internal"
$allPackages += Get-PackageDirs -BaseDir (Join-Path $repoRoot "cmd") -Prefix "cmd"

# Сортируем по importPath для детерминированного порядка.
$allPackages = $allPackages | Sort-Object ImportPath

# Записываем дамп без BOM (PowerShell 5.1 Set-Content -Encoding UTF8 добавляет BOM).
$dump = Join-Path $outDir "GODOC_DUMP.txt"
$utf8NoBom = [System.Text.UTF8Encoding]::new($false)
[System.IO.File]::WriteAllText($dump, "# godoc dump generated $(Get-Date -Format o)`n", $utf8NoBom)

Push-Location $repoRoot
try {
    $count = 0
    foreach ($pkg in $allPackages) {
        $header = "`n===== $($pkg.ImportPath) =====`n"
        [System.IO.File]::AppendAllText($dump, $header, $utf8NoBom)
        try {
            $output = & go doc -all $pkg.ImportPath 2>&1
            # В PS 5.1 stderr-записи приходят как ErrorRecord; извлекаем текст.
            $text = ($output | ForEach-Object { $_ -join "`n" }) -join "`n"
            [System.IO.File]::AppendAllText($dump, $text + "`n", $utf8NoBom)
            $count++
        } catch {
            [System.IO.File]::AppendAllText($dump, "[skip] error: $_`n", $utf8NoBom)
        }
    }
} finally {
    Pop-Location
}

Write-Host "==> Fallback dump written to $dump ($count packages)"
Write-Host "    Install godoc for full HTML: go install golang.org/x/tools/cmd/godoc@latest"
