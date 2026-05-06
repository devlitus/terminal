# SessionStart hook for the Architect agent.
# Reads the active task file from .github/tasks/in-progress/ and injects
# its content as a system message so the Architect starts every session
# already aware of the current scope, acceptance criteria, and tech notes.

try {
    $inProgress = Join-Path $PSScriptRoot '..\..\tasks\in-progress'
    $tasks = Get-ChildItem -Path $inProgress -Filter '*.md' -ErrorAction SilentlyContinue |
             Where-Object { $_.Name -ne '.gitkeep' }

    if ($tasks -and $tasks.Count -gt 0) {
        $task = $tasks | Sort-Object Name | Select-Object -First 1
        $content = Get-Content $task.FullName -Raw -ErrorAction SilentlyContinue

        if ($content) {
            $msg = "## Active Task: $($task.BaseName)`n`n$content"
            @{ systemMessage = $msg } | ConvertTo-Json -Compress
        }
    }
} catch {
    # On any error, do nothing — session continues normally
}

exit 0
