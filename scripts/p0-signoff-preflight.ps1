Param(
    [string]$Owner = "RekadzeAV",
    [string]$Repo = "network-scanner",
    [string]$WorkflowFile = "ci.yml"
)

$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $root

Write-Host "== P0 sign-off preflight (Windows) ==" -ForegroundColor Cyan

. (Join-Path $PSScriptRoot "resolve-github-token.ps1")
$hasToken = [bool]$env:GITHUB_TOKEN
$hasBash = $false
try {
    $bashCmd = Get-Command bash -ErrorAction SilentlyContinue
    if ($bashCmd) {
        & $bashCmd.Source -lc "true" *> $null
        $hasBash = ($LASTEXITCODE -eq 0)
    }
} catch {
    $hasBash = $false
}
if (-not $hasBash) {
    $gitBash = "C:\Program Files\Git\bin\bash.exe"
    if (Test-Path $gitBash) {
        try {
            & $gitBash -lc "true" *> $null
            $hasBash = ($LASTEXITCODE -eq 0)
        } catch {
            $hasBash = $false
        }
    }
}

$hasSh = $false
try {
    $shCmd = Get-Command sh -ErrorAction SilentlyContinue
    if ($shCmd) {
        & $shCmd.Source -lc "true" *> $null
        $hasSh = ($LASTEXITCODE -eq 0)
    }
} catch {
    $hasSh = $false
}
if (-not $hasSh) {
    $gitSh = "C:\Program Files\Git\usr\bin\sh.exe"
    if (Test-Path $gitSh) {
        try {
            & $gitSh -lc "true" *> $null
            $hasSh = ($LASTEXITCODE -eq 0)
        } catch {
            $hasSh = $false
        }
    }
}

Write-Host ""
Write-Host "Prerequisites:" -ForegroundColor Yellow
Write-Host ("- GITHUB_TOKEN: {0}" -f ($(if ($hasToken) { "SET" } else { "MISSING" })))
Write-Host ("- bash runtime: {0}" -f ($(if ($hasBash) { "FOUND" } else { "MISSING" })))
Write-Host ("- sh runtime: {0}" -f ($(if ($hasSh) { "FOUND" } else { "MISSING" })))

Write-Host ""
Write-Host "CI snapshot:" -ForegroundColor Yellow
& ".\scripts\check-ci-status.ps1" -Owner $Owner -Repo $Repo -WorkflowFile $WorkflowFile
$ciExitCode = $LASTEXITCODE
$ciRunsUri = "https://api.github.com/repos/$Owner/$Repo/actions/workflows/$WorkflowFile/runs?per_page=10"
$ciRuns = Invoke-RestMethod -Uri $ciRunsUri
$hasSuccessfulRun = @($ciRuns.workflow_runs | Where-Object { $_.status -eq "completed" -and $_.conclusion -eq "success" }).Count -ge 1
if ($ciExitCode -ne 0) {
    Write-Host ("- check-ci-status result: FAIL ({0})" -f $ciExitCode) -ForegroundColor Red
} else {
    Write-Host "- check-ci-status result: OK" -ForegroundColor Green
}
Write-Host ("- successful recent CI run: {0}" -f ($(if ($hasSuccessfulRun) { "YES" } else { "NO" })))

$failed = $false
if (-not $hasToken) {
    $failed = $true
    Write-Host "BLOCKER: GITHUB_TOKEN is not set." -ForegroundColor Red
}
if (-not $hasBash -and -not $hasSh) {
    $failed = $true
    Write-Host "BLOCKER: no bash/sh runtime for Unix closure scripts." -ForegroundColor Red
}
if ($ciExitCode -ne 0 -or -not $hasSuccessfulRun) {
    $failed = $true
    Write-Host "BLOCKER: no green CI evidence yet." -ForegroundColor Red
}

Write-Host ""
if ($failed) {
    Write-Host "Preflight result: BLOCKED" -ForegroundColor Red
    Write-Host "Runbook: docs/P0_SIGNOFF_RUNBOOK.md" -ForegroundColor Yellow
    exit 1
}

Write-Host "Preflight result: READY" -ForegroundColor Green
