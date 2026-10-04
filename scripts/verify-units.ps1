# verify-units.ps1 — проверка systemd-юнитов и desktop-файла (E7/7.6) на Windows.
# Парный скрипт к scripts/verify-units.sh: те же проверки для разработчиков на
# Windows и для CI-шага на windows-latest (bash там недоступен).
#
# Запуск: powershell -ExecutionPolicy Bypass -File .\scripts\verify-units.ps1

$ErrorActionPreference = 'Stop'

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RootDir = Split-Path -Parent $ScriptDir
$SystemdDir = Join-Path $RootDir 'config/systemd'
$DesktopFile = Join-Path $RootDir 'config/desktop/network-scanner-gui.desktop'

$script:Errors = 0
$script:Checks = 0

function Fail($msg) {
    $script:Errors++
    Write-Host "  [FAIL] $msg"
}

function Ok($msg) {
    $script:Checks++
    Write-Host "  [ok]   $msg"
}

function Get-KeyValue($path, $key) {
    if (-not (Test-Path $path)) { return '' }
    $match = Select-String -Path $path -Pattern "^$([regex]::Escape($key))=" | Select-Object -First 1
    if (-not $match) { return '' }
    return ($match.Line -split '=', 2)[1].Trim()
}

function Has-Section($path, $section) {
    if (-not (Test-Path $path)) { return $false }
    return [bool](Select-String -Path $path -Pattern "^\[$([regex]::Escape($section))\]$" -Quiet)
}

function Has-Key($path, $key) {
    if (-not (Test-Path $path)) { return $false }
    return [bool](Select-String -Path $path -Pattern "^$([regex]::Escape($key))=" -Quiet)
}

Write-Host '=== systemd unit validation ==='

$ServiceFile = Join-Path $SystemdDir 'network-scanner-scan.service'
$TimerFile = Join-Path $SystemdDir 'network-scanner-scan.timer'

# --- .service ---
if (-not (Test-Path $ServiceFile)) {
    Fail "missing $ServiceFile"
} else {
    Ok "found $(Split-Path -Leaf $ServiceFile)"

    foreach ($section in @('Unit', 'Service', 'Install')) {
        if (Has-Section $ServiceFile $section) { Ok "section [$section]" } else { Fail "missing section [$section]" }
    }

    foreach ($key in @('Description', 'Documentation', 'After', 'ExecStart', 'Type', 'User', 'Group')) {
        if (Has-Key $ServiceFile $key) { Ok "key $key" } else { Fail "missing key $key" }
    }

    $execStart = Get-KeyValue $ServiceFile 'ExecStart'
    if ($execStart -match '^/usr/local/bin/network-scanner\s+scan\b') {
        Ok 'ExecStart invokes installed binary with subcommand'
    } else {
        Fail "ExecStart must call /usr/local/bin/network-scanner scan: '$execStart'"
    }

    $type = Get-KeyValue $ServiceFile 'Type'
    if ($type -eq 'oneshot') { Ok 'Type=oneshot (разовое сканирование)' } else { Fail "Type should be oneshot, got '$type'" }

    if (Has-Key $ServiceFile 'ExecReload') {
        Fail 'ExecReload present but app has no SIGHUP handler'
    } else {
        Ok 'no bogus ExecReload'
    }

    $user = Get-KeyValue $ServiceFile 'User'
    if ($user -eq 'root') { Fail 'User=root: prefer dedicated user with capabilities' } else { Ok "dedicated user ($user)" }

    foreach ($key in @('NoNewPrivileges', 'ProtectSystem', 'ReadWritePaths')) {
        if (Has-Key $ServiceFile $key) { Ok $key } else { Fail "missing $key" }
    }

    $rw = Get-KeyValue $ServiceFile 'ReadWritePaths'
    if ($rw -like '*/var/lib/network-scanner*') {
        Ok 'ReadWritePaths includes /var/lib/network-scanner'
    } else {
        Fail "ReadWritePaths missing /var/lib/network-scanner: '$rw'"
    }

    $doc = Get-KeyValue $ServiceFile 'Documentation'
    if ($doc -like 'https://github.com/RekadzeAV/network-scanner*') {
        Ok 'Documentation points to the real repository'
    } else {
        Fail "Documentation URL must be https://github.com/RekadzeAV/network-scanner, got '$doc'"
    }
}
# --- .timer ---
if (-not (Test-Path $TimerFile)) {
    Fail "missing $TimerFile"
} else {
    Ok "found $(Split-Path -Leaf $TimerFile)"

    foreach ($section in @('Timer', 'Install')) {
        if (Has-Section $TimerFile $section) { Ok "section [$section]" } else { Fail "missing section [$section]" }
    }

    foreach ($key in @('OnBootSec', 'OnUnitActiveSec', 'Unit', 'WantedBy')) {
        if (Has-Key $TimerFile $key) { Ok "timer key $key" } else { Fail "missing timer key $key" }
    }

    $unitRef = Get-KeyValue $TimerFile 'Unit'
    if ($unitRef -and (Test-Path (Join-Path $SystemdDir $unitRef))) {
        Ok "timer references existing unit: $unitRef"
    } else {
        Fail "timer Unit='$unitRef' does not match any file in $SystemdDir"
    }
}

# --- desktop file ---
Write-Host '=== desktop entry validation ==='
if (-not (Test-Path $DesktopFile)) {
    Fail "missing $DesktopFile"
} else {
    Ok "found $(Split-Path -Leaf $DesktopFile)"

    $firstLine = (Get-Content $DesktopFile -TotalCount 1)
    if ($firstLine -eq '[Desktop Entry]') {
        Ok '[Desktop Entry] header present'
    } else {
        Fail 'first line must be [Desktop Entry]'
    }

    foreach ($key in @('Type', 'Name', 'Exec', 'Icon', 'TryExec')) {
        if (Has-Key $DesktopFile $key) { Ok "desktop key $key" } else { Fail "missing desktop key $key" }
    }

    $dExec = Get-KeyValue $DesktopFile 'Exec'
    $dTry = Get-KeyValue $DesktopFile 'TryExec'
    if ($dExec -eq '/usr/local/bin/network-scanner-gui') {
        Ok 'Exec points to installed GUI binary'
    } else {
        Fail "Exec should be /usr/local/bin/network-scanner-gui, got '$dExec'"
    }
    if ($dTry -eq $dExec) { Ok 'TryExec matches Exec' } else { Fail 'TryExec should match Exec' }

    # Приложение не регистрирует схему network:// — объявление даёт
    # неработающую ассоциацию в браузерах и меню «Открыть с помощью».
    $mime = Get-KeyValue $DesktopFile 'MimeType'
    if ($mime -like '*x-scheme-handler/network*') {
        Fail 'declares x-scheme-handler/network but app does not register the scheme'
    } else {
        Ok 'no unsupported scheme handler'
    }

    $icon = Get-KeyValue $DesktopFile 'Icon'
    if ($icon -and (Test-Path (Join-Path $RootDir "assets/icons/$icon.svg"))) {
        Ok "icon asset exists: assets/icons/$icon.svg"
    } else {
        Fail "icon asset assets/icons/$icon.svg not found"
    }
}

Write-Host '=== result ==='
Write-Host "checks passed: $script:Checks"
if ($script:Errors -gt 0) {
    Write-Host "errors: $script:Errors"
    exit 1
}
Write-Host 'errors: 0'
exit 0