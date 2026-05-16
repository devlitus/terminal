#!/usr/bin/env bash
# PostToolUse hook: runs go test -race when a _test.go file is written or edited.
# Outputs test results as systemMessage so the user sees them immediately.
# Claude Code docs: https://code.claude.com/docs/es/hooks

set -euo pipefail

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

# Only act on Go test files
if [[ "$FILE_PATH" != *_test.go ]]; then
  exit 0
fi

# Derive the package path from the file path
PKG_DIR=$(dirname "$FILE_PATH")
PKG_PATH="./$PKG_DIR/..."

TEST_OUTPUT=$(go test -race "$PKG_PATH" 2>&1 || true)
EXIT_CODE=$?

python3 -c "
import json, sys
output = sys.argv[1]
pkg = sys.argv[2]
code = int(sys.argv[3])
status = 'PASS' if code == 0 else 'FAIL'
print(json.dumps({
  'systemMessage': f'[TEST {status}] go test -race {pkg}:\n{output}'
}))
" "$TEST_OUTPUT" "$PKG_PATH" "$EXIT_CODE"

exit 0
