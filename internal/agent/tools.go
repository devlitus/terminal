package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/forge-tui/forge/internal/acp"
)

// Tool bundles an OpenAI tool definition with its Go implementation.
type Tool struct {
	Def     acp.ToolDef
	Execute func(ctx context.Context, args json.RawMessage) (string, error)
}

// defaultTools returns the MVP tool set. confirmFn is called before run_command
// executes; returning false cancels the execution.
func defaultTools(cwd func() string, confirmFn func(command string) bool) []Tool {
	return []Tool{
		runCommandTool(cwd, confirmFn),
		readFileTool(),
		listDirTool(),
		getCwdTool(cwd),
	}
}

// runCommandTool executes a shell command after asking the user for confirmation.
func runCommandTool(cwd func() string, confirmFn func(command string) bool) Tool {
	params := json.RawMessage(`{
		"type": "object",
		"properties": {
			"command": {
				"type": "string",
				"description": "The shell command to execute"
			}
		},
		"required": ["command"]
	}`)

	return Tool{
		Def: acp.ToolDef{
			Type: "function",
			Function: acp.FunctionDef{
				Name:        "run_command",
				Description: "Execute a shell command in the current working directory. Always asks the user for confirmation before running.",
				Parameters:  params,
			},
		},
		Execute: func(ctx context.Context, args json.RawMessage) (string, error) {
			var a struct {
				Command string `json:"command"`
			}
			if err := json.Unmarshal(args, &a); err != nil {
				return "", fmt.Errorf("run_command: invalid args: %w", err)
			}
			if a.Command == "" {
				return "", fmt.Errorf("run_command: empty command")
			}

			if !confirmFn(a.Command) {
				return "cancelled by user", nil
			}

			cmd := exec.CommandContext(ctx, "sh", "-c", a.Command)
			cmd.Dir = cwd()
			out, err := cmd.CombinedOutput()
			if err != nil {
				return strings.TrimSpace(string(out)), fmt.Errorf("exit error: %w", err)
			}
			return strings.TrimSpace(string(out)), nil
		},
	}
}

func readFileTool() Tool {
	params := json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {
				"type": "string",
				"description": "Absolute or relative path to the file"
			}
		},
		"required": ["path"]
	}`)

	return Tool{
		Def: acp.ToolDef{
			Type: "function",
			Function: acp.FunctionDef{
				Name:        "read_file",
				Description: "Read the contents of a file and return them as a string.",
				Parameters:  params,
			},
		},
		Execute: func(ctx context.Context, args json.RawMessage) (string, error) {
			var a struct {
				Path string `json:"path"`
			}
			if err := json.Unmarshal(args, &a); err != nil {
				return "", fmt.Errorf("read_file: invalid args: %w", err)
			}
			data, err := os.ReadFile(filepath.Clean(a.Path))
			if err != nil {
				return "", fmt.Errorf("read_file: %w", err)
			}
			return string(data), nil
		},
	}
}

func listDirTool() Tool {
	params := json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {
				"type": "string",
				"description": "Absolute or relative path to the directory. Defaults to current directory if empty."
			}
		},
		"required": []
	}`)

	return Tool{
		Def: acp.ToolDef{
			Type: "function",
			Function: acp.FunctionDef{
				Name:        "list_dir",
				Description: "List the contents of a directory.",
				Parameters:  params,
			},
		},
		Execute: func(ctx context.Context, args json.RawMessage) (string, error) {
			var a struct {
				Path string `json:"path"`
			}
			_ = json.Unmarshal(args, &a)
			if a.Path == "" {
				a.Path = "."
			}
			entries, err := os.ReadDir(filepath.Clean(a.Path))
			if err != nil {
				return "", fmt.Errorf("list_dir: %w", err)
			}
			var sb strings.Builder
			for _, e := range entries {
				if e.IsDir() {
					sb.WriteString(e.Name() + "/\n")
				} else {
					sb.WriteString(e.Name() + "\n")
				}
			}
			return strings.TrimRight(sb.String(), "\n"), nil
		},
	}
}

func getCwdTool(cwd func() string) Tool {
	params := json.RawMessage(`{"type": "object", "properties": {}, "required": []}`)

	return Tool{
		Def: acp.ToolDef{
			Type: "function",
			Function: acp.FunctionDef{
				Name:        "get_cwd",
				Description: "Get the current working directory.",
				Parameters:  params,
			},
		},
		Execute: func(ctx context.Context, args json.RawMessage) (string, error) {
			return cwd(), nil
		},
	}
}
