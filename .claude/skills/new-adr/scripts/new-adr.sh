#!/usr/bin/env bash
# Creates a numbered ADR file in .github/adr/ and prints the result.
# Usage: new-adr.sh <kebab-case-title>
# Bash equivalent of .github/skills/new-adr/scripts/new-adr.ps1

set -euo pipefail

TITLE="${1:-untitled}"

REPO_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
ADR_DIR="$REPO_ROOT/.github/adr"
mkdir -p "$ADR_DIR"

# Find the highest existing ADR number
MAX=0
for f in "$ADR_DIR"/ADR-*.md; do
  [[ -f "$f" ]] || continue
  NUM=$(basename "$f" | grep -oE '^ADR-[0-9]+' | grep -oE '[0-9]+' || echo 0)
  [[ "$NUM" -gt "$MAX" ]] && MAX="$NUM"
done

NEXT=$((MAX + 1))
PADDED=$(printf '%03d' "$NEXT")

# Normalize title to kebab-case
SAFE=$(echo "$TITLE" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g; s/-\+/-/g; s/^-//; s/-$//')

# Title case for the ADR header
TITLE_DISPLAY=$(echo "$TITLE" | sed 's/-/ /g' | awk '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) tolower(substr($i,2))}1')

FILE_NAME="ADR-$PADDED-$SAFE.md"
FILE_PATH="$ADR_DIR/$FILE_NAME"

cat > "$FILE_PATH" << TEMPLATE
# ADR-$PADDED: $TITLE_DISPLAY

## Status
Proposed

## Context
<!-- What problem exists? What forces are at play? Why does this decision matter now? -->

## Decision
<!-- What was decided? Be specific — name the pattern, technology, or boundary chosen. -->

## Consequences
**Easier**: <!-- What becomes simpler or better? -->
**Harder**: <!-- What becomes more complex or costly? -->
**Risks**: <!-- What could go wrong? What assumptions could be invalidated? -->
TEMPLATE

echo "FILE_PATH=$FILE_PATH"
echo "ADR_NUMBER=ADR-$PADDED"
echo "TITLE=$TITLE_DISPLAY"
