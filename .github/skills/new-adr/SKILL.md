---
name: new-adr
description: "Use when: creating an Architecture Decision Record (ADR), documenting an architectural decision, recording a design choice, or when the Architect says 'let me write an ADR for this'."
argument-hint: "Short kebab-case title, e.g. 'use-bubble-tea-v1'"
---

# new-adr

Creates a numbered ADR file in `.github/adr/` following the project's ADR format.

## When to Use

Invoke whenever an architectural decision needs to be recorded — after a trade-off analysis, a pattern selection, or a boundary definition. The Architect should call this skill before closing any significant design discussion.

## Steps

1. Run [new-adr.ps1](./scripts/new-adr.ps1) with the kebab-case title as argument to determine the next ADR number and create the file.
2. Read the created file path from the script output.
3. Open the file and fill in the **Context**, **Decision**, and **Consequences** sections based on the current discussion.
4. Report the file path and ADR number to the user.

## ADR Format

```markdown
# ADR-NNN: [Title in sentence case]

## Status
Proposed

## Context
[What problem exists? What forces are at play? Why does this decision matter now?]

## Decision
[What was decided? Be specific — name the pattern, technology, or boundary chosen.]

## Consequences
**Easier**: [What becomes simpler or better?]
**Harder**: [What becomes more complex or costly?]
**Risks**: [What could go wrong? What assumptions could be invalidated?]
```

## Notes

- ADR numbers are zero-padded to 3 digits: `ADR-001`, `ADR-002`, …
- File names follow: `ADR-NNN-kebab-case-title.md`
- Status starts as `Proposed`; the Orchestrator updates it to `Accepted` after review
- Never renumber existing ADRs — supersede them with a new one
