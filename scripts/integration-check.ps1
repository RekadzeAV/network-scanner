Param()

$ErrorActionPreference = "Stop"

if (Get-Command docker-compose -ErrorAction SilentlyContinue) {
    Write-Host "docker-compose detected. Starting integration stack."
    docker-compose up -d
    $startedCompose = $true
} else {
    Write-Host "docker-compose not found. Running local smoke checks only."
    $startedCompose = $false
}

try {
    .\scripts\smoke-cli-no-topology.ps1
    .\scripts\smoke-cli-topology.ps1
    Write-Host "Integration check passed."
}
finally {
    if ($startedCompose) {
        docker-compose down
    }
}
