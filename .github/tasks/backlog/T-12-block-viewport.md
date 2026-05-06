---
id: T-12
status: backlog
agent: Code Expert
depends-on: [T-08]
branch: T-12/block-viewport
---

## T-12 · Block Viewport

**Context**
The scrollable list of block components that occupies the main body of the TUI. It manages focus, j/k navigation, and is the render container for all visible blocks.

**Acceptance criteria**
- [ ] `internal/ui/viewport/model.go` defines `Model` implementing `tea.Model`
- [ ] Holds a slice of `blockui.Model`; renders all visible blocks stacked vertically in `View()`
- [ ] `j` / `down` moves focus to the next (newer) block; `k` / `up` moves to the previous (older) block; stops at bounds, does not wrap
- [ ] The focused block index is tracked; `View()` sets `focused=true` on the correct block component via a `SetFocused(bool)` method
- [ ] On `tea.WindowSizeMsg`, recalculates available height and propagates width to all child block models
- [ ] Auto-scrolls to the newest block when a new block is appended (unless the user has manually scrolled up)
- [ ] Emits `BlockFocusedMsg { BlockID uint64 }` when the focused block changes
- [ ] `go test ./internal/ui/viewport/...` passes: focus navigation, focus-change message emission, new block auto-scroll

**Tech notes**
- FR-06, FR-07, FR-08 (PRD §6.2)
- Use `github.com/charmbracelet/bubbles/viewport` as the scroll container wrapping the stacked block views
- Import `internal/ui/block` as `blockui` to avoid name collision
- Available height = terminal height − 2 (header 1 line + input bar 1 line)
- `SetFocused(bool)` is a method added to `blockui.Model` — define it in T-08 if not already present

**Completion notes**
*(empty)*
