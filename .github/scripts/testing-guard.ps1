# PreToolUse hook for the Testing agent.
# Denies any attempt to edit a file that is not a *_test.go file.
# Copilot pipes a JSON object on stdin describing the tool call.

try {
    $json = [System.Console]::In.ReadToEnd()
    $data = $json | ConvertFrom-Json

    $editTools = @(
        'replace_string_in_file',
        'multi_replace_string_in_file',
        'create_file',
        'edit_file',
        'write_file'
    )

    $toolName = if ($data.PSObject.Properties['tool_name']) { $data.tool_name }
                elseif ($data.PSObject.Properties['toolName']) { $data.toolName }
                else { '' }

    if ($editTools -contains $toolName) {
        $toolInput = if ($data.PSObject.Properties['tool_input']) { $data.tool_input }
                     elseif ($data.PSObject.Properties['toolInput']) { $data.toolInput }
                     else { $null }

        $filePath = ''
        if ($toolInput) {
            if ($toolInput.PSObject.Properties['filePath'])   { $filePath = $toolInput.filePath }
            elseif ($toolInput.PSObject.Properties['file_path']) { $filePath = $toolInput.file_path }
            elseif ($toolInput.PSObject.Properties['path'])   { $filePath = $toolInput.path }
        }

        if ($filePath -and $filePath -notmatch '_test\.go$') {
            @{
                hookSpecificOutput = @{
                    hookEventName            = 'PreToolUse'
                    permissionDecision       = 'deny'
                    permissionDecisionReason = "Testing agent may only edit *_test.go files. '$filePath' is a source file — read it, do not modify it."
                }
            } | ConvertTo-Json -Compress
            exit 2
        }
    }
} catch {
    # On any parsing error, allow the tool to proceed
}

exit 0
