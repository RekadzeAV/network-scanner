param(
    [int]$TimeoutSec = 1
)

$ErrorActionPreference = "Stop"
$PSNativeCommandUseErrorActionPreference = $false

function Assert-LastExitCode {
    param([string]$Step)
    if ($LASTEXITCODE -ne 0) {
        throw "$Step failed with exit code $LASTEXITCODE"
    }
}

function Get-NodeIdentity {
    param(
        [string]$Hostname,
        [string]$IP,
        [string]$MAC
    )
    foreach ($v in @($Hostname, $IP, $MAC)) {
        if ($null -eq $v) { continue }
        $t = $v.ToString().Trim().ToLower()
        if ($t -ne "") { return $t }
    }
    return ""
}

function Get-UndirectedEdgeKey {
    param(
        [string]$A,
        [string]$B
    )
    $a = $A.Trim().ToLower()
    $b = $B.Trim().ToLower()
    if ($a -le $b) { return "$a<->$b" }
    return "$b<->$a"
}

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $root

Write-Host "== Smoke: D-track topology export ==" -ForegroundColor Cyan

$tmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ("network-scanner-dtrack-" + [guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $tmpDir | Out-Null

$smokeExe = Join-Path $tmpDir "network-scanner-smoke.exe"
go build -o $smokeExe .\cmd\network-scanner
Assert-LastExitCode "go build smoke binary"

$baseArgs = @("--network", "127.0.0.1/32", "--timeout", "$TimeoutSec", "--ports", "1-8", "--topology")
$jsonOut = Join-Path $tmpDir "topology.json"
$graphmlOut = Join-Path $tmpDir "topology.graphml"
$pngOut = Join-Path $tmpDir "topology.png"
$pngLog = Join-Path $tmpDir "png.log"

try {
    & $smokeExe @baseArgs --output-format json --output-file $jsonOut *> $null
    Assert-LastExitCode "json export"
    & $smokeExe @baseArgs --output-format graphml --output-file $graphmlOut *> $null
    Assert-LastExitCode "graphml export"

    if (!(Test-Path $jsonOut)) { throw "JSON export file not found" }
    if (!(Test-Path $graphmlOut)) { throw "GraphML export file not found" }

    $json = Get-Content -Path $jsonOut -Raw
    $graphml = Get-Content -Path $graphmlOut -Raw
    if ($json -notmatch '"Devices"') { throw "JSON export misses Devices field" }
    if ($graphml -notmatch "<graphml") { throw "GraphML export misses graphml root" }

    $jsonObj = $json | ConvertFrom-Json
    $graphmlObj = [xml]$graphml

    $jsonNodes = New-Object System.Collections.Generic.List[string]
    foreach ($p in $jsonObj.Devices.PSObject.Properties) {
        $d = $p.Value
        $jsonNodes.Add((Get-NodeIdentity -Hostname $d.Hostname -IP $d.IP -MAC $d.MAC))
    }
    $jsonNodes = $jsonNodes | Sort-Object

    $labelById = @{}
    $graphmlNodes = New-Object System.Collections.Generic.List[string]
    foreach ($n in $graphmlObj.graphml.graph.node) {
        $label = ""
        foreach ($d in $n.data) {
            if ($d.key -eq "label") {
                $label = $d.'#text'.ToString().Trim().ToLower()
                break
            }
        }
        $labelById[$n.id.ToString().Trim()] = $label
        if ($label -ne "") { $graphmlNodes.Add($label) }
    }
    $graphmlNodes = $graphmlNodes | Sort-Object

    if ((@($jsonNodes) -join "|") -ne (@($graphmlNodes) -join "|")) {
        throw "Node set mismatch between JSON and GraphML"
    }

    $jsonEdges = New-Object System.Collections.Generic.List[string]
    foreach ($l in $jsonObj.Links) {
        $src = Get-NodeIdentity -Hostname $l.Source.Hostname -IP $l.Source.IP -MAC $l.Source.MAC
        $dst = Get-NodeIdentity -Hostname $l.Target.Hostname -IP $l.Target.IP -MAC $l.Target.MAC
        $jsonEdges.Add((Get-UndirectedEdgeKey -A $src -B $dst))
    }
    $jsonEdges = $jsonEdges | Sort-Object

    $graphmlEdges = New-Object System.Collections.Generic.List[string]
    foreach ($e in $graphmlObj.graphml.graph.edge) {
        $src = $labelById[$e.source.ToString().Trim()]
        $dst = $labelById[$e.target.ToString().Trim()]
        $graphmlEdges.Add((Get-UndirectedEdgeKey -A $src -B $dst))
    }
    $graphmlEdges = $graphmlEdges | Sort-Object

    if ((@($jsonEdges) -join "|") -ne (@($graphmlEdges) -join "|")) {
        throw "Edge set mismatch between JSON and GraphML"
    }

    $oldEap = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    & $smokeExe @baseArgs --output-format png --output-file $pngOut *>&1 | Out-File -FilePath $pngLog -Encoding utf8
    $pngExit = $LASTEXITCODE
    $ErrorActionPreference = $oldEap
    if ($pngExit -ne 0) {
        throw "png export failed with exit code $pngExit"
    }

    if (Test-Path $pngOut) {
        Write-Host "PNG export produced image via Graphviz." -ForegroundColor Yellow
    }
    elseif (Test-Path $jsonOut) {
        $log = Get-Content -Path $pngLog -Raw
        if ($log -notmatch "Graphviz недоступен|fallback JSON") {
            throw "Fallback JSON exists but expected message missing"
        }
        Write-Host "PNG export fallback to JSON works when dot is unavailable." -ForegroundColor Yellow
    }
    else {
        throw "Neither PNG nor fallback JSON was produced"
    }
}
finally {
    Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
}

Write-Host "Smoke passed: D-track topology exports are healthy." -ForegroundColor Green
