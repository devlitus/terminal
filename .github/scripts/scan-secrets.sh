#!/usr/bin/env bash
# Scans files read by the Security Expert agent for hardcoded secrets patterns.
# Called as PostToolUse hook from security.agent.md

set -euo pipefail

HOOK_INPUT=$(cat)

TOOL_NAME=$(echo "$HOOK_INPUT" | python3 -c "
import sys, json
d = json.load(sys.stdin)
print(d.get('tool', ''))
" 2>/dev/null || echo "")

# Only scan on file read operations
if [[ "$TOOL_NAME" != "read_file" && "$TOOL_NAME" != "str_replace_editor" ]]; then
  exit 0
fi

FILE_PATH=$(echo "$HOOK_INPUT" | python3 -c "
import sys, json
d = json.load(sys.stdin)
ti = d.get('toolInput', {})
print(ti.get('path', '') or ti.get('filePath', ''))
" 2>/dev/null || echo "")

if [[ -z "$FILE_PATH" || ! -f "$FILE_PATH" ]]; then
  exit 0
fi

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

FINDINGS=""
for pattern in "${PATTERNS[@]}"; do
  MATCHES=$(grep -nEi "$pattern" "$FILE_PATH" 2>/dev/null || true)
  if [[ -n "$MATCHES" ]]; then
    FINDINGS+="$MATCHES"$'\n'
  fi
done

if [[ -n "$FINDINGS" ]]; then
  echo ""
  echo "WARNING: [SECRET SCANNER] Potential hardcoded secrets detected in: $FILE_PATH"
  echo "---"
  echo "$FINDINGS"
  echo "---"
  echo "Review these lines before proceeding with the analysis."
  echo ""
fi

exit 0
