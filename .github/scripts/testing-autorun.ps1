# PostToolUse hook for the Testing agent.
# After editing a *_test.go file, runs go test on that package and injects
# the result as a system message so the agent can see failures immediately.

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
            if ($toolInput.PSObject.Properties['filePath'])      { $filePath = $toolInput.filePath }
            elseif ($toolInput.PSObject.Properties['file_path']) { $filePath = $toolInput.file_path }
            elseif ($toolInput.PSObject.Properties['path'])      { $filePath = $toolInput.path }
        }

        if ($filePath -match '_test\.go$') {
            $dir = Split-Path $filePath -Parent
            $result = & go test $dir 2>&1 | Out-String
            @{
                systemMessage = "go test result:`n$result"
            } | ConvertTo-Json -Compress
        }
    }
} catch {
    # On any parsing error, do nothing
}

exit 0
