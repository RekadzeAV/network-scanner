param(
    [switch]$SkipTests,
    [switch]$NoColor
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# --- Цвета -------------------------------------------------------------------

function Write-OK  { param([string]$Text); Write-Host "[OK]  $Text"  -ForegroundColor Green }
function Write-FAIL { param([string]$Text); Write-Host "[FAIL] $Text" -ForegroundColor Red }
function Write-INFO { param([string]$Text); Write-Host "[--]  $Text" -ForegroundColor Cyan }
function Write-SEP  { param([string]$Text); Write-Host "=== $Text ===" -ForegroundColor Magenta }

# --- Go installed? -----------------------------------------------------------

$goCmd = Get-Command go -ErrorAction SilentlyContinue
if (-not $goCmd) { throw 'go not found in PATH' }

# --- Root --------------------------------------------------------------------

$Root = $PWD.Path
$GuiDir = Join-Path $Root 'internal\gui'

# --- Counters ----------------------------------------------------------------

$TotalChecks  = 0
$PassedChecks = 0

function Report-OK  { param([string]$Name); Write-OK $Name; $script:TotalChecks++; $script:PassedChecks++ }
function Report-FAIL { param([string]$Name); Write-FAIL $Name; $script:TotalChecks++ }

# ============================================================================
# (a) hardcoded palette literals — only theme_tokens.go
#
# Правило §13 по смыслу запрещает ЖЁСТКО ЗАШИТЫЕ ЗНАЧЕНИЯ палитры вне
# theme_tokens.go, а не сам синтаксис color.RGBA{...}. Поэтому проверяется
# содержимое литерала, а не факт его наличия:
#
#   FAIL  — R/G/B инициализированы числовым литералом (R: 0x12, R: 18) либо
#           литерал позиционный и содержит числа (color.RGBA{255, 0, 0, 255});
#   OK    — значения вычисляются (byte(r), mix(c.R), byte((val >> 24) & 0xFF)),
#           литерал пуст (color.RGBA{} как нулевое значение ошибки) или
#           встречается в комментарии.
#
# Альфа-канал (A:) намеренно не проверяется: A: 0xFF — идиома «непрозрачный»,
# а не цвет палитры.
#
# Известное ограничение: комментарии вырезаются эвристикой («//» до конца
# строки без учёта строковых литералов). Ложных срабатываний это не даёт,
# теоретически может пропустить литерал внутри raw-строки с «//» — в GUI-коде
# таких мест нет.
# ============================================================================
Write-SEP 'Point (a): hardcoded color.RGBA literals'

# Блоковые комментарии /* */ вырезаются с сохранением длин (смещения строк
# остаются валидными), чтобы литерал внутри комментария не считался кодом.
function Remove-BlockComments {
    param([string]$Text)
    return [regex]::Replace($Text, '/\*[\s\S]*?\*/', { param($m) (' ' * $m.Value.Length) })
}

# Строчные комментарии //... вырезаются до конца строки (длина сохраняется).
function Remove-LineComments {
    param([string]$Text)
    return [regex]::Replace($Text, '(?m)//[^\r\n]*', { param($m) (' ' * $m.Value.Length) })
}

# Номер строки по смещению в тексте (1-индексированный).
function Get-LineNumber {
    param([string]$Text, [int]$Index)
    return ($Text.Substring(0, $Index).Split("`n").Count)
}

$failCountA = 0
# Ключевой синтаксис с числовым литералом в канале цвета: R: 0x12 / G: 34 / B: 0
$reKeyedNumeric = '(?<![A-Za-z0-9_])[RGB]\s*:\s*(?:0[xX])?[0-9]+\s*[,}]'
# Позиционный синтаксис: внутри есть цифры, но нет ни одного ключа канала.
$reAnyDigit = '[0-9]'
$reAnyKey = '(?<![A-Za-z0-9_])[RGB]\s*:'

$goFiles = Get-ChildItem -Path $GuiDir -Filter '*.go' -Recurse
foreach ($f in $goFiles) {
    $name = $f.Name
    if ($name -match '_test\.go$') { continue }
    if ($name -eq 'theme_tokens.go') { continue }

    $text = [IO.File]::ReadAllText($f.FullName)
    if ($text -notmatch 'color\.RGBA') { continue }

    $code = Remove-BlockComments -Text $text
    $code = Remove-LineComments -Text $code

    $rel = $f.FullName.Substring($Root.Length).TrimStart('\')

    foreach ($m in [regex]::Matches($code, 'color\.RGBA\s*\{[^{}]*\}', 'Singleline')) {
        $body = $m.Value

        $isKeyed = [regex]::IsMatch($body, $reAnyKey)
        $violation = $false
        if ($isKeyed) {
            if ([regex]::IsMatch($body, $reKeyedNumeric)) { $violation = $true }
        }
        elseif ([regex]::IsMatch($body, $reAnyDigit)) {
            # Позиционный литерал с числами — это зашитое значение цвета.
            $violation = $true
        }

        if ($violation) {
            $failCountA++
            $lineNo = Get-LineNumber -Text $code -Index $m.Index
            $snippet = ($body -replace '\s+', ' ').Trim()
            if ($snippet.Length -gt 72) { $snippet = $snippet.Substring(0, 69) + '...' }
            Write-FAIL "  $rel : $lineNo  $snippet"
        }
    }
}

if ($failCountA -eq 0) {
    Report-OK 'no hardcoded color.RGBA palette literals outside theme_tokens.go'
}
else {
    Report-FAIL "hardcoded color.RGBA palette literals outside theme_tokens.go ($failCountA violation(s) found)"
}

# ============================================================================
# (b) gofmt -l internal/gui
# ============================================================================
Write-SEP 'Point (b): gofmt'

Set-Location $Root
$gofmtOut = gofmt -l internal\gui 2>&1

$gofmtClean = $true
if ($null -ne $gofmtOut -and "$gofmtOut".Trim() -ne '') {
    $gofmtClean = $false
}

if ($gofmtClean) {
    Report-OK 'gofmt -l internal\gui -- clean'
}
else {
    Report-FAIL "gofmt -l internal\gui -- not formatted: $gofmtOut"
}

# ============================================================================
# (c) go vet
# ============================================================================
Write-SEP 'Point (c): go vet'

$vetOut = go vet ./internal/gui/... 2>&1
$vetOk  = ($LASTEXITCODE -eq 0)

if ($vetOk) {
    Report-OK 'go vet ./internal/gui/...'
}
else {
    Report-FAIL "go vet ./internal/gui/...: $vetOut"
}

# ============================================================================
# (d) go test ./internal/gui/... -short
# ============================================================================
Write-SEP 'Point (d): go test ./internal/gui/...'

if ($SkipTests) {
    Write-INFO 'Point (d) skipped (-SkipTests)'
}
else {
    $testOut = go test ./internal/gui/... -count=1 -short 2>&1
    $testOk  = ($LASTEXITCODE -eq 0)

    if ($testOk) {
        Report-OK 'go test ./internal/gui/... -count=1 -short'
    }
    else {
        Report-FAIL "go test ./internal/gui/... -count=1 -short: $testOut"
    }
}

# ============================================================================
# (e) TestWCAGSweep
# ============================================================================
Write-SEP 'Point (e): go test TestWCAGSweep'

if ($SkipTests) {
    Write-INFO 'Point (e) skipped (-SkipTests)'
}
else {
    $wcagOut = go test ./internal/gui/ -run TestWCAGSweep -count=1 -short 2>&1
    $wcagOk  = ($LASTEXITCODE -eq 0)

    if ($wcagOk) {
        Report-OK 'go test -run TestWCAGSweep -count=1 -short'
    }
    else {
        Report-FAIL "go test -run TestWCAGSweep -count=1 -short: $wcagOut"
    }
}

# ============================================================================
# Summary
# ============================================================================
Write-SEP 'SUMMARY'

$failed = $TotalChecks - $PassedChecks

if ($failed -eq 0) {
    Write-OK "All $TotalChecks checks passed"
    exit 0
}
else {
    Write-FAIL "$failed of $TotalChecks checks FAILED"
    exit 1
}
