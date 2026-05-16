#!/usr/bin/env bash
# PostToolUse hook: runs go build + go vet when a .go file is written or edited.
# Outputs systemMessage with build errors so Claude can fix them immediately.
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

# Only act on Go source files
if [[ "$FILE_PATH" != *.go ]]; then
  exit 0
fi

BUILD_OUTPUT=$(go build ./... 2>&1 || true)
VET_OUTPUT=$(go vet ./... 2>&1 || true)

ERRORS=""
[[ -n "$BUILD_OUTPUT" ]] && ERRORS+="go build:\n$BUILD_OUTPUT\n"
[[ -n "$VET_OUTPUT" ]] && ERRORS+="go vet:\n$VET_OUTPUT\n"

if [[ -n "$ERRORS" ]]; then
  python3 -c "
import json, sys
errors = sys.argv[1]
print(json.dumps({
  'systemMessage': f'[BUILD] Errors after editing {sys.argv[2]}:\n{errors}',
  'additionalContext': f'Build/vet failed after editing {sys.argv[2]}. Errors:\n{errors}\nFix these before proceeding.'
}))
" "$ERRORS" "$FILE_PATH"
fi

exit 0
