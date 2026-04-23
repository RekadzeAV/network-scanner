Param()

$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $root

Write-Host "== Stage2 sign-off status (Windows) ==" -ForegroundColor Cyan

function Invoke-StepCheck {
    param(
        [string]$Name,
        [scriptblock]$Action
    )

    Write-Host ""
    Write-Host ("[STEP] {0}" -f $Name) -ForegroundColor Yellow
    $null = & $Action
    $code = $LASTEXITCODE
    if ($code -eq 0) {
        Write-Host ("[OK] {0}" -f $Name) -ForegroundColor Green
        return $true
    }

    Write-Host ("[FAIL] {0} (exit={1})" -f $Name, $code) -ForegroundColor Red
    return $false
}

$results = @{}

$results["stage2-p1"] = [bool](Invoke-StepCheck -Name "Stage2 P1 closure" -Action { powershell -ExecutionPolicy Bypass -File ".\scripts\stage2-p1-closure-check.ps1" })
$results["stage2-p2"] = [bool](Invoke-StepCheck -Name "Stage2 P2 closure" -Action { powershell -ExecutionPolicy Bypass -File ".\scripts\stage2-p2-closure-check.ps1" })
$results["stage2-p3"] = [bool](Invoke-StepCheck -Name "Stage2 P3 closure" -Action { powershell -ExecutionPolicy Bypass -File ".\scripts\stage2-p3-closure-check.ps1" })
$results["docs-links"] = [bool](Invoke-StepCheck -Name "Docs local link sanity" -Action { powershell -ExecutionPolicy Bypass -File ".\scripts\docs-link-check.ps1" })
$results["p0-preflight"] = [bool](Invoke-StepCheck -Name "P0 sign-off preflight" -Action { powershell -ExecutionPolicy Bypass -File ".\scripts\p0-signoff-preflight.ps1" })

$passed = @($results.GetEnumerator() | Where-Object { $_.Value -eq $true }).Count
$total = $results.Count

Write-Host ""
Write-Host "== Summary ==" -ForegroundColor Cyan
$results.GetEnumerator() | Sort-Object Name | ForEach-Object {
    $mark = if ($_.Value -eq $true) { "PASS" } else { "FAIL" }
    Write-Host ("- {0}: {1}" -f $_.Key, $mark)
}

Write-Host ("Result: {0}/{1} checks passed." -f $passed, $total)
if ($passed -eq $total) {
    Write-Host "Stage2 sign-off status: READY" -ForegroundColor Green
    exit 0
}

Write-Host "Stage2 sign-off status: BLOCKED" -ForegroundColor Red
Write-Host "See: docs/RELEASE_READY_GAP_LIST.md and docs/P0_SIGNOFF_RUNBOOK.md" -ForegroundColor Yellow
exit 1
