Param(
    [string]$RootPath = "."
)

$ErrorActionPreference = "Stop"

$root = Resolve-Path $RootPath
Set-Location $root

Write-Host "== Docs local link check ==" -ForegroundColor Cyan

$mdFiles = Get-ChildItem -Path $root -Recurse -File -Include *.md
$broken = New-Object System.Collections.Generic.List[string]

$linkRegex = [regex]'\[[^\]]+\]\(([^)#]+)'

foreach ($file in $mdFiles) {
    $content = Get-Content -Path $file.FullName -Raw -Encoding UTF8
    $linkMatches = $linkRegex.Matches($content)
    foreach ($m in $linkMatches) {
        $link = $m.Groups[1].Value.Trim()
        if ($link -match '^(https?:|mailto:)') {
            continue
        }

        # Decode URL-encoded paths to support non-ASCII filenames.
        $decoded = [System.Uri]::UnescapeDataString($link)
        $target = [System.IO.Path]::GetFullPath((Join-Path $file.DirectoryName $decoded))
        if (-not (Test-Path -LiteralPath $target)) {
            $relFile = Resolve-Path -LiteralPath $file.FullName -Relative
            $broken.Add("$relFile -> $link")
        }
    }
}

if ($broken.Count -gt 0) {
    Write-Host "Broken local markdown links found:" -ForegroundColor Red
    $broken | ForEach-Object { Write-Host "- $_" }
    exit 1
}

Write-Host "OK: no broken local markdown links." -ForegroundColor Green
