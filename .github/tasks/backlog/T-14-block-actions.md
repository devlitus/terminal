---
id: T-14
status: backlog
agent: Code Expert
depends-on: [T-12, T-13]
branch: T-14/block-actions
---

## T-14 · Block Actions — Copy and Re-run

**Context**
Adds `y` (copy to clipboard) and `r` (re-run) keyboard shortcuts to the focused block. These are the two non-AI block interactions defined in FR-09 and FR-10.

**Acceptance criteria**
- [ ] `y` on a focused, non-running block copies `block.Output` joined with `\n` to the system clipboard
- [ ] `r` on a focused block submits the same `Command` again via the existing shell execution flow (T-13), creating a new block
- [ ] Clipboard write uses `github.com/atotto/clipboard` — add to `go.mod`
- [ ] If clipboard is unavailable (e.g., no `$DISPLAY`), `y` appends a one-line notice `"clipboard unavailable"` in ink-6 to the viewport status area — does not crash
- [ ] `r` on a `StateRunning` block is a no-op (no duplicate execution, no error)
- [ ] `y` and `r` are no-ops when no block is focused
- [ ] `go test ./...` continues to pass after these changes

**Tech notes**
- FR-09, FR-10 (PRD §6.2)
- Key handling lives at the root model or viewport level, not inside `blockui.Model`
- `github.com/atotto/clipboard` is cgo-free and cross-platform — prefer it over `golang.design/x/clipboard`
- `r` reuses the exact same code path as `input.SubmitMsg` — it should emit or directly call the same handler

**Completion notes**
*(empty)*
