Param(
    [string]$Owner = "RekadzeAV",
    [string]$Repo = "network-scanner",
    [string]$WorkflowFile = "ci.yml",
    [string]$ChecklistPath = "docs/P3_CLOSURE_CHECKLIST.md",
    [string]$ConfirmedBy = "TBD",
    [string]$Date = ""
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
    return ($lintOk -and $testOk -and $buildOk)
}

if ([string]::IsNullOrWhiteSpace($Date)) {
    $Date = Get-Date -Format "yyyy-MM-dd"
}

$runsUri = "https://api.github.com/repos/$Owner/$Repo/actions/workflows/$WorkflowFile/runs?per_page=20"
$resp = Invoke-RestMethod -Uri $runsUri
$ok = $resp.workflow_runs | Where-Object { $_.status -eq "completed" -and $_.conclusion -eq "success" } | Select-Object -First 1
if ($null -eq $ok) {
    throw "No successful CI run found for workflow '$WorkflowFile'."
}

$jobs = Get-RunJobs -OwnerName $Owner -RepoName $Repo -RunId $ok.id
if (-not (Test-RequiredJobsGreen -Jobs $jobs)) {
    throw "Latest successful run does not satisfy required jobs: Lint/Test*/Build and Smoke*."
}

$fullChecklistPath = Resolve-Path $ChecklistPath
$lines = [System.Collections.Generic.List[string]](Get-Content -LiteralPath $fullChecklistPath)
$bt = [char]96

function Set-FirstBacktickValue {
    param(
        [string]$Line,
        [string]$Value
    )
    if ($Line -match "^(.*?`)[^`]*(`.*)$") {
        return ($matches[1] + $Value + $matches[2])
    }
    return $Line
}

function Set-TextAfterFirstBacktickPair {
    param(
        [string]$Line,
        [string]$Tail
    )
    if ($Line -match "^(.*?`[^`]*`)(.*)$") {
        return ($matches[1] + $Tail)
    }
    return $Line
}

$sectionStart = -1
for ($i = 0; $i -lt $lines.Count; $i++) {
    if ($lines[$i] -eq "## P3 Final Sign-off") {
        $sectionStart = $i
        break
    }
}
if ($sectionStart -lt 0) {
    throw "P3 Final Sign-off section not found in checklist."
}

$sectionEnd = $lines.Count - 1
for ($i = $sectionStart + 1; $i -lt $lines.Count; $i++) {
    if ($lines[$i] -match "^## ") {
        $sectionEnd = $i - 1
        break
    }
}

$bulletIdx = @()
for ($i = $sectionStart; $i -le $sectionEnd; $i++) {
    if ($lines[$i].TrimStart().StartsWith("- ")) {
        $bulletIdx += $i
    }
}
if ($bulletIdx.Count -lt 8) {
    throw "Unexpected P3 Final Sign-off format. Not enough bullet items."
}

$lines[$bulletIdx[0]] = Set-FirstBacktickValue -Line $lines[$bulletIdx[0]] -Value "closed"
$lines[$bulletIdx[1]] = Set-FirstBacktickValue -Line $lines[$bulletIdx[1]] -Value $Date
$lines[$bulletIdx[2]] = Set-FirstBacktickValue -Line $lines[$bulletIdx[2]] -Value $ConfirmedBy
$lines[$bulletIdx[3]] = Set-FirstBacktickValue -Line $lines[$bulletIdx[3]] -Value $ok.html_url
$lines[$bulletIdx[6]] = Set-FirstBacktickValue -Line $lines[$bulletIdx[6]] -Value "closed"
$lines[$bulletIdx[6]] = Set-TextAfterFirstBacktickPair -Line $lines[$bulletIdx[6]] -Tail " - required jobs (`Lint`, `Test*`, `Build and Smoke*`) confirmed green."
$lines[$bulletIdx[7]] = Set-FirstBacktickValue -Line $lines[$bulletIdx[7]] -Value ".\scripts\finalize-p3-signoff.ps1"
$lines[$bulletIdx[7]] = Set-TextAfterFirstBacktickPair -Line $lines[$bulletIdx[7]] -Tail (" - run " + $bt + $ok.id + $bt + " validated.")

Set-Content -LiteralPath $fullChecklistPath -Value $lines -Encoding UTF8

Write-Host "P3 checklist updated from successful CI run." -ForegroundColor Green
Write-Host ("Run URL: {0}" -f $ok.html_url)
