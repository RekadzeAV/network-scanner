Param(
    [string]$Owner = "RekadzeAV",
    [string]$Repo = "network-scanner",
    [string]$WorkflowFile = "ci.yml",
    [string]$Ref = "main",
    [string]$ConfirmedBy = "TBD",
    [int]$TimeoutMinutes = 30,
    [int]$PollSeconds = 15
)

$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $root

Write-Host "== P3 close all (Windows) ==" -ForegroundColor Cyan

. (Join-Path $PSScriptRoot "resolve-github-token.ps1")
if (-not $env:GITHUB_TOKEN) {
    throw @"
GITHUB_TOKEN is not set. Use one of:
  - `$env:GITHUB_TOKEN = '<token>'` in this session, or
  - persist for your user: [Environment]::SetEnvironmentVariable('GITHUB_TOKEN','<token>','User'), or
  - install GitHub CLI and run: gh auth login
Then re-run p3-close-all.
"@
}
if (-not (Get-Command powershell -ErrorAction SilentlyContinue)) {
    throw "PowerShell runtime is required."
}

Write-Host "[1/3] Trigger CI workflow and wait for completion" -ForegroundColor Yellow
& ".\scripts\trigger-ci-workflow.ps1" `
    -Owner $Owner `
    -Repo $Repo `
    -WorkflowFile $WorkflowFile `
    -Ref $Ref `
    -TimeoutMinutes $TimeoutMinutes `
    -PollSeconds $PollSeconds

Write-Host "[2/3] Check latest successful CI status" -ForegroundColor Yellow
& ".\scripts\check-ci-status.ps1" `
    -Owner $Owner `
    -Repo $Repo `
    -WorkflowFile $WorkflowFile

Write-Host "[3/3] Finalize P3 sign-off in checklist" -ForegroundColor Yellow
& ".\scripts\finalize-p3-signoff.ps1" `
    -Owner $Owner `
    -Repo $Repo `
    -WorkflowFile $WorkflowFile `
    -ConfirmedBy $ConfirmedBy

Write-Host "P3 close-all flow completed successfully." -ForegroundColor Green
