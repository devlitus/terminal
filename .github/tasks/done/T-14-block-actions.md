---
id: T-14
status: done
agent: Code Expert
depends-on: [T-12, T-13]
branch: T-14/block-actions
---

## T-14 · Block Actions — Copy and Re-run

**Context**
Adds `y` (copy to clipboard) and `r` (re-run) keyboard shortcuts to the focused block. These are the two non-AI block interactions defined in FR-09 and FR-10.

**Acceptance criteria**
- [x] `y` on a focused, non-running block copies `block.Output` joined with `\n` to the system clipboard
- [x] `r` on a focused block submits the same `Command` again via the existing shell execution flow (T-13), creating a new block
- [x] Clipboard write uses `github.com/atotto/clipboard` — add to `go.mod`
- [x] If clipboard is unavailable (e.g., no `$DISPLAY`), `y` appends a one-line notice `"clipboard unavailable"` in ink-6 to the viewport status area — does not crash
- [x] `r` on a `StateRunning` block is a no-op (no duplicate execution, no error)
- [x] `y` and `r` are no-ops when no block is focused
- [x] `go test ./...` continues to pass after these changes

**Tech notes**
- FR-09, FR-10 (PRD §6.2)
- Key handling lives at the root model or viewport level, not inside `blockui.Model`
- `github.com/atotto/clipboard` is cgo-free and cross-platform — prefer it over `golang.design/x/clipboard`
- `r` reuses the exact same code path as `input.SubmitMsg` — it should emit or directly call the same handler

**Completion notes**
- `Block()` getter added to `internal/ui/block/model.go` to expose block data to the viewport.
- `y`/`r` key handlers added to `internal/ui/viewport/model.go`; `statusNotice` field drives the clipboard-unavailable notice styled in ink-6.
- `github.com/atotto/clipboard` promoted from indirect to direct dependency in `go.mod`.
- Five new tests added to `model_test.go`: `TestCopyNoFocus`, `TestRerunNoFocus`, `TestRerunRunningBlock`, `TestRerunIdleBlock`, `TestCopyGracefulOnUnavailable`.
