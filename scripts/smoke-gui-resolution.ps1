param(
    [string]$GuiExe = ""
)

$ErrorActionPreference = "Stop"

function Resolve-GuiExe {
    param([string]$InputPath)
    if ($InputPath -and (Test-Path $InputPath)) {
        return (Resolve-Path $InputPath).Path
    }
    $defaultPath = Join-Path (Resolve-Path (Join-Path $PSScriptRoot "..")).Path "network-scanner-gui.exe"
    if (Test-Path $defaultPath) {
        return $defaultPath
    }
    throw "GUI binary not found. Pass -GuiExe <path-to-network-scanner-gui.exe>."
}

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $root

$exe = Resolve-GuiExe -InputPath $GuiExe
Write-Host "== Smoke: GUI resolution matrix ==" -ForegroundColor Cyan
Write-Host "Binary: $exe"
Write-Host ""
Write-Host "Run checks manually for each profile:"
Write-Host "  1) 1366x768 @100% and @125%"
Write-Host "  2) 1920x1080 @125%"
Write-Host "  3) 2560x1440 @100%"
Write-Host "  4) 3840x2160 @150%"
Write-Host ""
Write-Host "Acceptance checklist:"
Write-Host "  - Scan tab: controls visible, no critical clipping."
Write-Host "  - Results tab: table/cards usable, Host Details actions accessible."
Write-Host "  - Topology tab: preview and text both reachable in windowed mode."
Write-Host "  - Tools tab: tool buttons and Operations controls available."
Write-Host "  - Same functionality in windowed and fullscreen modes."
Write-Host ""
Write-Host "Launching GUI..."
Start-Process -FilePath $exe | Out-Null
Write-Host "Close the GUI after verification and record findings in release notes." -ForegroundColor Green
