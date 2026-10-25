param([string]$Root = ".")

$activeDocs = @(
    "README.md",
    "docs/README.md",
    "docs/ARCHITECTURE.md",
    "docs/TECHNICAL.md",
    "docs/USER_GUIDE.md",
    "docs/GUI.md",
    "docs/PROJECT_STRUCTURE.md",
    "docs/ROADMAP.md",
    "docs/UNIFIED_OPTIMIZED_PLAN_2026-09-15.md",
    "docs/THREE_PLANS_ANALYSIS_2026-09-15.md",
    "docs/FINAL_PLAN_2026-09-15.md",
    "docs/CLI_REFERENCE.md",
    "docs/INSTALL.md",
    "docs/BUILD_STRUCTURE.md",
    "docs/LOGGING.md",
    "docs/deployment.md",
    "docs/GUI_SMOKE_CHECKLIST.md",
    "docs/GRAPHML_COMPATIBILITY_CHECK.md"
)

$pattern = "\[([^\]]+)\]\(([^)]+)\)"
$broken = @()

foreach ($doc in $activeDocs) {
    if (-not (Test-Path $doc)) { continue }
    $docDir = Split-Path -Parent $doc
    if ([string]::IsNullOrWhiteSpace($docDir)) { $docDir = "." }
    $content = Get-Content $doc -Raw -ErrorAction SilentlyContinue
    if ($null -eq $content) { continue }
    [regex]::Matches($content, $pattern) | ForEach-Object {
        $link = $_.Groups[2].Value
        if ([string]::IsNullOrWhiteSpace($link)) { return }
        if ($link -match "^http" -or $link -match "^#" -or $link -match "^mailto:") { return }
        $linkDecoded = [System.Uri]::UnescapeDataString($link)
        $target = Join-Path $docDir $linkDecoded
        $target = $target -replace "/", [System.IO.Path]::DirectorySeparatorChar
        if (-not (Test-Path $target)) {
            $broken += "BROKEN in ${doc}: $link -> $target"
        }
    }
}

if ($broken.Count -gt 0) {
    Write-Host "Найдено битых ссылок: $($broken.Count)"
    $broken | ForEach-Object { Write-Host $_ }
    exit 1
} else {
    Write-Host "Все ссылки в активных документах целы."
    exit 0
}
