#!/usr/bin/env bash
# Scans a file for hardcoded secrets.
#
# Hook mode (no args): PostToolUse hook — reads JSON from stdin, only acts on Write|Edit.
# Direct mode (file path as arg): scans the given file and prints findings to stdout.
#   Usage: bash scan-secrets.sh <file-path>
#
# Claude Code docs: https://code.claude.com/docs/es/hooks

set -euo pipefail

PATTERNS=(
  'password[[:space:]]*=[[:space:]]*["'"'"'][^"'"'"']{4,}'
  'passwd[[:space:]]*=[[:space:]]*["'"'"'][^"'"'"']{4,}'
  'secret[[:space:]]*=[[:space:]]*["'"'"'][^"'"'"']{4,}'
  'api_key[[:space:]]*=[[:space:]]*["'"'"'][^"'"'"']{4,}'
  'apikey[[:space:]]*=[[:space:]]*["'"'"'][^"'"'"']{4,}'
  'access_token[[:space:]]*=[[:space:]]*["'"'"'][^"'"'"']{4,}'
  'auth_token[[:space:]]*=[[:space:]]*["'"'"'][^"'"'"']{4,}'
  'private_key[[:space:]]*=[[:space:]]*["'"'"'][^"'"'"']{4,}'
  'AKIA[0-9A-Z]{16}'
  'AIza[0-9A-Za-z_-]{35}'
  'ghp_[0-9a-zA-Z]{36}'
  'gho_[0-9a-zA-Z]{36}'
  'sk-[a-zA-Z0-9]{32,}'
  'BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY'
)

scan_file() {
  local file="$1"
  local findings=""

  for pattern in "${PATTERNS[@]}"; do
    local matches
    matches=$(grep -nEi "$pattern" "$file" 2>/dev/null || true)
    [[ -n "$matches" ]] && findings+="$matches"$'\n'
  done

  echo "$findings"
}

# --- Direct mode: file path passed as argument ---
if [[ $# -gt 0 ]]; then
  FILE_PATH="$1"
  if [[ ! -f "$FILE_PATH" ]]; then
    echo "File not found: $FILE_PATH" >&2
    exit 1
  fi
  FINDINGS=$(scan_file "$FILE_PATH")
  if [[ -n "$FINDINGS" ]]; then
    echo "[SECRET SCANNER] Potential hardcoded secrets in $FILE_PATH:"
    echo "$FINDINGS"
  else
    echo "[SECRET SCANNER] No secrets found in $FILE_PATH."
  fi
  exit 0
fi

# --- Hook mode: JSON payload via stdin ---
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

if [[ -z "$FILE_PATH" || ! -f "$FILE_PATH" ]]; then
  exit 0
fi

FINDINGS=$(scan_file "$FILE_PATH")

if [[ -n "$FINDINGS" ]]; then
  python3 -c "
import json, sys
path, findings = sys.argv[1], sys.argv[2]
print(json.dumps({
  'systemMessage': f'[SECRET SCANNER] Potential hardcoded secrets in {path}:\n{findings}\nReview before committing.'
}))
" "$FILE_PATH" "$FINDINGS"
fi

exit 0
