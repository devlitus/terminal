---
id: T-08
status: backlog
agent: Code Expert
depends-on: [T-03, T-04]
branch: T-08/block-component
---

## T-08 · Command Block UI Component

**Context**
The Bubble Tea model that renders a single command block — the atom of the Forge UI. Must faithfully implement the block anatomy from PRD §10 and handle all three states (running, success, failed) with correct design tokens.

**Acceptance criteria**
- [ ] `internal/ui/block/model.go` defines `Model` implementing `tea.Model` (`Init`, `Update`, `View`)
- [ ] `View()` renders three stacked sections: header line (bg=ink-2, cwd + duration + status badge), input line (bg=ink-1, ember-500 `❯` prompt, command in ink-9), output section (bg=ink-0, ink-8 text)
- [ ] Output section is capped at `terminal-height / 3` lines; when output exceeds the cap, the last N lines are shown (tail behaviour)
- [ ] Focused state wraps the block in `theme.BlockBorderFocused` (ember-500 rounded border); unfocused uses `theme.BlockBorderDefault`
- [ ] `BadgeSuccess` renders with mint-500 background, `BadgeFailed` with crimson-500, `BadgeRunning` with solar-500
- [ ] Component renders without panic at both 80-column and 220-column terminal widths
- [ ] Width is passed via a `SetWidth(n int)` method on `Model`; the component uses it for all `lipgloss` width constraints
- [ ] `go test ./internal/ui/block/...` passes: correct badge text per state, focused/unfocused border styles differ

**Tech notes**
- PRD §10 block anatomy and §5.1 block states
- `Model` wraps a `block.Block` value (not a pointer) — it renders from the data, does not mutate it
- Import `internal/theme` for all styles and colors; never hard-code hex values in this package
- Import `internal/block` for `Block` and `BlockState` types
- The AI card (T-11) is rendered separately by the parent — this component only renders the shell block sections
- `lipgloss.NewStyle().MaxHeight(n).MaxWidth(w)` for output section constraints

**Completion notes**
*(empty)*
