Param(
    [string]$Owner = "RekadzeAV",
    [string]$Repo = "network-scanner",
    [string]$WorkflowFile = "ci.yml",
    [string]$Ref = "main",
    [int]$TimeoutMinutes = 30,
    [int]$PollSeconds = 15
)

$ErrorActionPreference = "Stop"

function Get-RunJobs {
    param(
        [string]$OwnerName,
        [string]$RepoName,
        [string]$RunId,
        [hashtable]$RequestHeaders
    )
    $jobsUri = "https://api.github.com/repos/$OwnerName/$RepoName/actions/runs/$RunId/jobs?per_page=100"
    $jobsResp = Invoke-RestMethod -Uri $jobsUri -Headers $RequestHeaders
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

if (-not $env:GITHUB_TOKEN) {
    throw "GITHUB_TOKEN is not set. Create a token with repo/workflow permissions and export it before running."
}

$headers = @{
    "Accept"               = "application/vnd.github+json"
    "Authorization"        = "Bearer $($env:GITHUB_TOKEN)"
    "X-GitHub-Api-Version" = "2022-11-28"
}

$dispatchUri = "https://api.github.com/repos/$Owner/$Repo/actions/workflows/$WorkflowFile/dispatches"
$payload = @{
    ref = $Ref
} | ConvertTo-Json

Write-Host "Dispatching workflow '$WorkflowFile' on ref '$Ref'..." -ForegroundColor Cyan
Invoke-RestMethod -Method Post -Uri $dispatchUri -Headers $headers -Body $payload -ContentType "application/json"

$deadline = (Get-Date).AddMinutes($TimeoutMinutes)
$runUri = "https://api.github.com/repos/$Owner/$Repo/actions/workflows/$WorkflowFile/runs?per_page=20"
$selected = $null

Write-Host "Waiting for the run to appear..." -ForegroundColor Yellow
while ((Get-Date) -lt $deadline -and $null -eq $selected) {
    $resp = Invoke-RestMethod -Uri $runUri -Headers $headers
    $selected = $resp.workflow_runs | Where-Object { $_.head_branch -eq $Ref } | Select-Object -First 1
    if ($null -eq $selected) {
        Start-Sleep -Seconds $PollSeconds
    }
}

if ($null -eq $selected) {
    throw "Timed out while waiting for a new workflow run to appear."
}

Write-Host ("Run detected: id={0} status={1}" -f $selected.id, $selected.status) -ForegroundColor Cyan
Write-Host ("URL: {0}" -f $selected.html_url)
Write-Host "Waiting for completion..." -ForegroundColor Yellow

while ((Get-Date) -lt $deadline) {
    $run = Invoke-RestMethod -Uri ("https://api.github.com/repos/$Owner/$Repo/actions/runs/$($selected.id)") -Headers $headers
    Write-Host ("status={0} conclusion={1} updated={2}" -f $run.status, $run.conclusion, $run.updated_at)
    if ($run.status -eq "completed") {
        Write-Host ""
        Write-Host ("Final run URL: {0}" -f $run.html_url)
        if ($run.conclusion -eq "success") {
            $jobs = Get-RunJobs -OwnerName $Owner -RepoName $Repo -RunId $run.id -RequestHeaders $headers
            $required = Test-RequiredJobsGreen -Jobs $jobs
            Write-Host "Required jobs check for P3 closure:" -ForegroundColor Cyan
            Write-Host ("- Lint: {0}" -f ($(if ($required.LintOk) { "OK" } else { "FAIL" })))
            Write-Host ("- Test matrix: {0} ({1} jobs)" -f ($(if ($required.TestOk) { "OK" } else { "FAIL" }), $required.TestTotal))
            Write-Host ("- Build and Smoke matrix: {0} ({1} jobs)" -f ($(if ($required.BuildOk) { "OK" } else { "FAIL" }), $required.BuildTotal))
            Write-Host ("- Stage2 P1 Closure: {0} ({1} jobs)" -f ($(if ($required.Stage2P1Ok) { "OK" } else { "FAIL" }), $required.Stage2P1Total))
            Write-Host ("- Stage2 P3 Closure: {0} ({1} jobs)" -f ($(if ($required.Stage2P3Ok) { "OK" } else { "FAIL" }), $required.Stage2P3Total))
            if (-not $required.AllRequired) {
                throw "Workflow is success, but required jobs (Lint/Test/Build and Smoke/Stage2 closures) are not fully green."
            }
            Write-Host "CI completed successfully; required jobs are green." -ForegroundColor Green
            exit 0
        }
        throw ("CI completed with conclusion '{0}'." -f $run.conclusion)
    }
    Start-Sleep -Seconds $PollSeconds
}

throw "Timed out waiting for workflow completion."
