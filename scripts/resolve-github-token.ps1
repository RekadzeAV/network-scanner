# Populates $env:GITHUB_TOKEN when unset: process (already set), User/Machine
# registry env, then `gh auth token` if GitHub CLI is installed.
if (-not [string]::IsNullOrWhiteSpace($env:GITHUB_TOKEN)) {
    return
}

foreach ($scope in @("User", "Machine")) {
    $t = [Environment]::GetEnvironmentVariable("GITHUB_TOKEN", $scope)
    if (-not [string]::IsNullOrWhiteSpace($t)) {
        $env:GITHUB_TOKEN = $t.Trim()
        return
    }
}

$ghExe = $null
$ghCmd = Get-Command gh -ErrorAction SilentlyContinue
if ($ghCmd -and (Test-Path -LiteralPath $ghCmd.Source)) {
    $ghExe = $ghCmd.Source
}
if (-not $ghExe) {
    $candidates = @(
        "${env:ProgramFiles}\GitHub CLI\gh.exe",
        "${env:ProgramFiles(x86)}\GitHub CLI\gh.exe"
    )
    foreach ($c in $candidates) {
        if ($c -and (Test-Path -LiteralPath $c)) {
            $ghExe = $c
            break
        }
    }
}

if ($ghExe) {
    try {
        $t = & $ghExe auth token 2>$null
        if (-not [string]::IsNullOrWhiteSpace($t)) {
            $env:GITHUB_TOKEN = $t.Trim()
        }
    } catch {
        # ignore: gh not logged in or token scope unavailable
    }
}
