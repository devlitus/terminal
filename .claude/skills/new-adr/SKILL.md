---
name: new-adr
description: "Creates a numbered Architecture Decision Record in .github/adr/. Use when documenting an architectural decision, pattern selection, package boundary, or significant design choice. The Architect agent should invoke this before closing any significant design discussion."
argument-hint: "kebab-case-title (e.g. use-bubble-tea-v1)"
allowed-tools: Read, Write
---

# new-adr: $ARGUMENTS

## File created by numbering script

!`bash "${CLAUDE_SKILL_DIR}/scripts/new-adr.sh" "$ARGUMENTS"`

## Your job

The script above created the ADR file and printed `FILE_PATH`, `ADR_NUMBER`, and `TITLE`.

1. Open the file at the path shown in `FILE_PATH`.
2. Fill in the three sections based on the current conversation:
   - **Context**: what problem exists, what forces are at play, why this matters now.
   - **Decision**: what was decided — name the specific pattern, technology, or boundary chosen.
   - **Consequences**: what becomes easier, what becomes harder, what risks remain.
3. Remove all `<!-- ... -->` placeholder comments.
4. Report back: file path, ADR number, and a one-line summary of the decision.

## Rules

- Status stays `Proposed` — the Orchestrator updates it to `Accepted` after review.
- Never renumber existing ADRs — supersede them with a new one.
- Keep the Decision section specific: vague decisions ("we chose a good approach") are useless.
