#!/usr/bin/env bash
# PostToolUse hook: runs golangci-lint on the affected package when a .go file is edited.
# Exits silently if golangci-lint is not installed.
# Claude Code docs: https://code.claude.com/docs/es/hooks

set -euo pipefail

command -v golangci-lint &>/dev/null || exit 0

HOOK_INPUT=$(cat)

TOOL_NAME=$(echo "$HOOK_INPUT" | python3 -c "
import sys, json
d = json.load(sys.stdin)
print(d.get('tool_name', ''))
" 2>/dev/null || echo "")

case "$TOOL_NAME" in
  Write|Edit) ;;
  *) exit 0 ;;
esac

FILE_PATH=$(echo "$HOOK_INPUT" | python3 -c "
import sys, json
d = json.load(sys.stdin)
ti = d.get('tool_input', {})
print(ti.get('file_path', '') or ti.get('path', ''))
" 2>/dev/null || echo "")

if [[ "$FILE_PATH" != *.go ]]; then
  exit 0
fi

PKG_DIR=$(dirname "$FILE_PATH")
PKG_PATH="./$PKG_DIR/..."

set +e
LINT_OUTPUT=$(golangci-lint run "$PKG_PATH" 2>&1)
EXIT_CODE=$?
set -e

if [[ $EXIT_CODE -ne 0 && -n "$LINT_OUTPUT" ]]; then
  python3 -c "
import json, sys
output, pkg = sys.argv[1], sys.argv[2]
print(json.dumps({
  'systemMessage': f'[LINT] golangci-lint {pkg}:\n{output}'
}))
" "$LINT_OUTPUT" "$PKG_PATH"
fi

exit 0
