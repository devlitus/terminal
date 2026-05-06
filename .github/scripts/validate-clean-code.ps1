# validate-clean-code.ps1
# PostToolUse hook: runs after file edits to enforce code quality.
# Receives tool JSON on stdin. Extracts the edited file path and runs
# the appropriate linter if available. Exits with code 2 to block on errors.

param()

# Read hook input from stdin
$hookInput = $null
try {
    $raw = [Console]::In.ReadToEnd()
    if ($raw) { $hookInput = $raw | ConvertFrom-Json }
} catch { }

# Only act on file edit tools
$toolName = $hookInput?.toolName
if ($toolName -notin @('edit', 'replace_string_in_file', 'multi_replace_string_in_file', 'create_file')) {
    exit 0
}

# Extract the file path from the tool input
$filePath = $hookInput?.toolInput?.filePath
if (-not $filePath -or -not (Test-Path $filePath)) {
    exit 0
}

$ext = [System.IO.Path]::GetExtension($filePath).ToLower()
$errors = @()

# --- JavaScript / TypeScript: ESLint ---
if ($ext -in @('.js', '.ts', '.jsx', '.tsx', '.mjs', '.cjs')) {
    $eslint = Get-Command eslint -ErrorAction SilentlyContinue
    if ($eslint) {
        $result = & eslint --format compact $filePath 2>&1
        if ($LASTEXITCODE -ne 0) {
            $errors += "ESLint: $result"
        }
    }
}

# --- Python: flake8 or pylint ---
if ($ext -eq '.py') {
    $flake8 = Get-Command flake8 -ErrorAction SilentlyContinue
    if ($flake8) {
        $result = & flake8 --max-line-length=100 $filePath 2>&1
        if ($LASTEXITCODE -ne 0) {
            $errors += "flake8: $result"
        }
    } else {
        $pylint = Get-Command pylint -ErrorAction SilentlyContinue
        if ($pylint) {
            $result = & pylint --output-format=text $filePath 2>&1
            if ($LASTEXITCODE -ge 4) {  # pylint exits 4+ on errors
                $errors += "pylint: $result"
            }
        }
    }
}

# --- C#: dotnet format ---
if ($ext -eq '.cs') {
    $dotnet = Get-Command dotnet -ErrorAction SilentlyContinue
    if ($dotnet) {
        $result = & dotnet format --verify-no-changes --diagnostics 2>&1
        if ($LASTEXITCODE -ne 0) {
            $errors += "dotnet format: $result"
        }
    }
}

# --- Java: checkstyle (if jar exists) ---
if ($ext -eq '.java') {
    $checkstyle = Get-ChildItem -Path . -Filter "checkstyle*.jar" -Recurse -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($checkstyle) {
        $result = & java -jar $checkstyle.FullName -c /google_checks.xml $filePath 2>&1
        if ($LASTEXITCODE -ne 0) {
            $errors += "checkstyle: $result"
        }
    }
}

# --- Output result ---
if ($errors.Count -gt 0) {
    $message = "Code quality check failed after edit:`n" + ($errors -join "`n")
    @{
        decision      = "block"
        stopReason    = $message
        systemMessage = $message
    } | ConvertTo-Json -Compress
    exit 2
}

exit 0
