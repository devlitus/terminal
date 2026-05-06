# Scans files read by the Security Expert agent for hardcoded secrets patterns.
# Called as PostToolUse hook from security.agent.md

param()

try {
    $hookInput = [Console]::In.ReadToEnd() | ConvertFrom-Json -ErrorAction Stop
} catch {
    exit 0
}

$toolName = $hookInput.tool
if ($toolName -ne "read_file" -and $toolName -ne "str_replace_editor") {
    exit 0
}

$filePath = $hookInput.toolInput.path
if (-not $filePath) { $filePath = $hookInput.toolInput.filePath }
if (-not $filePath -or -not (Test-Path $filePath -PathType Leaf)) {
    exit 0
}

$patterns = @(
    'password\s*=\s*["''][^"'']{4,}',
    'passwd\s*=\s*["''][^"'']{4,}',
    'secret\s*=\s*["''][^"'']{4,}',
    'api_key\s*=\s*["''][^"'']{4,}',
    'apikey\s*=\s*["''][^"'']{4,}',
    'access_token\s*=\s*["''][^"'']{4,}',
    'auth_token\s*=\s*["''][^"'']{4,}',
    'private_key\s*=\s*["''][^"'']{4,}',
    'AKIA[0-9A-Z]{16}',
    'AIza[0-9A-Za-z\-_]{35}',
    'ghp_[0-9a-zA-Z]{36}',
    'gho_[0-9a-zA-Z]{36}',
    'sk-[a-zA-Z0-9]{32,}',
    'BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY'
)

$findings = @()
$content = Get-Content $filePath -ErrorAction SilentlyContinue
if (-not $content) { exit 0 }

for ($i = 0; $i -lt $content.Count; $i++) {
    foreach ($p in $patterns) {
        if ($content[$i] -match $p) {
            $findings += "  Line $($i + 1): $($content[$i].Trim())"
            break
        }
    }
}

if ($findings.Count -gt 0) {
    Write-Output ""
    Write-Output "WARNING: [SECRET SCANNER] Potential hardcoded secrets detected in: $filePath"
    Write-Output "---"
    $findings | ForEach-Object { Write-Output $_ }
    Write-Output "---"
    Write-Output "Review these lines before proceeding with the analysis."
    Write-Output ""
}

exit 0
