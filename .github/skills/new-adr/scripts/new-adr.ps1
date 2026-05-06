# new-adr.ps1
# Usage: .\new-adr.ps1 <kebab-case-title>
# Creates the next numbered ADR file in .github/adr/ and prints the file path.

param(
    [Parameter(Mandatory = $true)]
    [string]$Title
)

$repoRoot = git -C $PSScriptRoot rev-parse --show-toplevel 2>$null
if (-not $repoRoot) {
    # Fallback: navigate up from scripts/ → new-adr/ → skills/ → .github/ → repo root
    $repoRoot = Resolve-Path (Join-Path $PSScriptRoot '..\..\..\..')
}

$adrDir = Join-Path $repoRoot '.github\adr'
if (-not (Test-Path $adrDir)) {
    New-Item -ItemType Directory -Path $adrDir | Out-Null
}

# Find the highest existing ADR number
$existing = Get-ChildItem -Path $adrDir -Filter 'ADR-*.md' -ErrorAction SilentlyContinue
$maxNum = 0
foreach ($f in $existing) {
    if ($f.Name -match '^ADR-(\d+)') {
        $n = [int]$Matches[1]
        if ($n -gt $maxNum) { $maxNum = $n }
    }
}

$nextNum   = $maxNum + 1
$paddedNum = $nextNum.ToString('D3')
$safeName  = $Title.ToLower() -replace '[^a-z0-9]+', '-' -replace '^-|-$', ''
$fileName  = "ADR-$paddedNum-$safeName.md"
$filePath  = Join-Path $adrDir $fileName

$titleCase = (Get-Culture).TextInfo.ToTitleCase($Title.Replace('-', ' '))

$template = @"
# ADR-$paddedNum`: $titleCase

## Status
Proposed

## Context
<!-- What problem exists? What forces are at play? Why does this decision matter now? -->

## Decision
<!-- What was decided? Be specific — name the pattern, technology, or boundary chosen. -->

## Consequences
**Easier**: <!-- What becomes simpler or better? -->
**Harder**: <!-- What becomes more complex or costly? -->
**Risks**: <!-- What could go wrong? What assumptions could be invalidated? -->
"@

Set-Content -Path $filePath -Value $template -Encoding UTF8
Write-Output $filePath
