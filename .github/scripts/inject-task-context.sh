#!/usr/bin/env bash
# SessionStart hook: injects active task context into Claude's session via additionalContext.
# Claude Code docs: https://code.claude.com/docs/es/hooks

set -euo pipefail

TASKS_DIR=".github/tasks/in-progress"

python3 - << 'PYEOF'
import json, os, sys

tasks_dir = ".github/tasks/in-progress"
lines = []

try:
    files = sorted(f for f in os.listdir(tasks_dir) if f.startswith("T-") and f.endswith(".md"))
    if not files:
        print(json.dumps({
            "additionalContext": "No tasks in-progress. Check .github/tasks/index.md for the backlog."
        }))
        sys.exit(0)

    lines.append("=== Active Tasks ===")
    for fname in files:
        path = os.path.join(tasks_dir, fname)
        with open(path) as f:
            content = f.read(1500)
        lines.append(f"\n--- {path} ---")
        lines.append(content)
except FileNotFoundError:
    print(json.dumps({"additionalContext": f"Task directory '{tasks_dir}' not found."}))
    sys.exit(0)
except Exception as e:
    print(json.dumps({"additionalContext": f"Could not read tasks: {e}"}))
    sys.exit(0)

print(json.dumps({"additionalContext": "\n".join(lines)}))
PYEOF
