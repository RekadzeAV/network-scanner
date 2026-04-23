Param(
    [string]$Owner = "RekadzeAV",
    [string]$Repo = "network-scanner",
    [string]$WorkflowFile = "ci.yml",
    [int]$Limit = 10
)

$ErrorActionPreference = "Stop"

function Get-RunJobs {
    param(
        [string]$OwnerName,
        [string]$RepoName,
        [string]$RunId
    )
    $jobsUri = "https://api.github.com/repos/$OwnerName/$RepoName/actions/runs/$RunId/jobs?per_page=100"
    $jobsResp = Invoke-RestMethod -Uri $jobsUri
    return $jobsResp.jobs
}

function Test-RequiredJobsGreen {
    param([array]$Jobs)

    $lintOk = @($Jobs | Where-Object { $_.name -eq "Lint" -and $_.conclusion -eq "success" }).Count -ge 1
    $testTotal = @($Jobs | Where-Object { $_.name -like "Test*" }).Count
    $testOk = @($Jobs | Where-Object { $_.name -like "Test*" -and $_.conclusion -eq "success" }).Count -eq $testTotal -and $testTotal -ge 3
    $buildTotal = @($Jobs | Where-Object { $_.name -like "Build and Smoke*" }).Count
    $buildOk = @($Jobs | Where-Object { $_.name -like "Build and Smoke*" -and $_.conclusion -eq "success" }).Count -eq $buildTotal -and $buildTotal -ge 3
    $stage2P1Total = @($Jobs | Where-Object { $_.name -like "Stage2 P1 Closure*" }).Count
    $stage2P1Ok = @($Jobs | Where-Object { $_.name -like "Stage2 P1 Closure*" -and $_.conclusion -eq "success" }).Count -eq $stage2P1Total -and $stage2P1Total -ge 1
    $stage2P3Total = @($Jobs | Where-Object { $_.name -like "Stage2 P3 Closure*" }).Count
    $stage2P3Ok = @($Jobs | Where-Object { $_.name -like "Stage2 P3 Closure*" -and $_.conclusion -eq "success" }).Count -eq $stage2P3Total -and $stage2P3Total -ge 1

    [PSCustomObject]@{
        LintOk      = $lintOk
        TestOk      = $testOk
        BuildOk     = $buildOk
        Stage2P1Ok  = $stage2P1Ok
        Stage2P3Ok  = $stage2P3Ok
        TestTotal   = $testTotal
        BuildTotal  = $buildTotal
        Stage2P1Total = $stage2P1Total
        Stage2P3Total = $stage2P3Total
        AllRequired = ($lintOk -and $testOk -and $buildOk -and $stage2P1Ok -and $stage2P3Ok)
    }
}

$uri = "https://api.github.com/repos/$Owner/$Repo/actions/workflows/$WorkflowFile/runs?per_page=$Limit"
$resp = Invoke-RestMethod -Uri $uri

if ($null -eq $resp.workflow_runs -or $resp.workflow_runs.Count -eq 0) {
    Write-Host "No CI runs found." -ForegroundColor Yellow
    exit 0
}

Write-Host "Recent CI runs ($Owner/$Repo, workflow=$WorkflowFile):" -ForegroundColor Cyan
$resp.workflow_runs | ForEach-Object {
    Write-Host ("- id={0} status={1} conclusion={2} updated={3}" -f $_.id, $_.status, $_.conclusion, $_.updated_at)
    Write-Host ("  " + $_.html_url)
}

$ok = $resp.workflow_runs | Where-Object { $_.status -eq "completed" -and $_.conclusion -eq "success" } | Select-Object -First 1
if ($null -eq $ok) {
    Write-Host "No successful run found in recent history." -ForegroundColor Yellow
    exit 0
}

Write-Host ""
Write-Host "Latest successful run:" -ForegroundColor Green
Write-Host ("id={0}" -f $ok.id)
Write-Host ("url={0}" -f $ok.html_url)

$jobs = Get-RunJobs -OwnerName $Owner -RepoName $Repo -RunId $ok.id
$required = Test-RequiredJobsGreen -Jobs $jobs

Write-Host ""
Write-Host "Required jobs check for P3 closure:" -ForegroundColor Cyan
Write-Host ("- Lint: {0}" -f ($(if ($required.LintOk) { "OK" } else { "FAIL" })))
Write-Host ("- Test matrix: {0} ({1} jobs)" -f ($(if ($required.TestOk) { "OK" } else { "FAIL" }), $required.TestTotal))
Write-Host ("- Build and Smoke matrix: {0} ({1} jobs)" -f ($(if ($required.BuildOk) { "OK" } else { "FAIL" }), $required.BuildTotal))
Write-Host ("- Stage2 P1 Closure: {0} ({1} jobs)" -f ($(if ($required.Stage2P1Ok) { "OK" } else { "FAIL" }), $required.Stage2P1Total))
Write-Host ("- Stage2 P3 Closure: {0} ({1} jobs)" -f ($(if ($required.Stage2P3Ok) { "OK" } else { "FAIL" }), $required.Stage2P3Total))
Write-Host ("- All required jobs green: {0}" -f ($(if ($required.AllRequired) { "YES" } else { "NO" })))
if (-not $required.AllRequired) {
    exit 1
}
