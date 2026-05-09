---
id: T-17
status: backlog
agent: Code Expert
depends-on: [T-10, T-12]
branch: T-17/command-palette
---

## T-17 · Command Palette

**Context**
The `ctrl+k` fuzzy-search modal over recent commands and built-in actions. It is a distinct overlay component that must not be conflated with the input bar or viewport (FR-19, FR-20).

**Acceptance criteria**
- [ ] `internal/ui/palette/model.go` defines `Model` implementing `tea.Model`
- [ ] Opens when the root model receives `OpenPaletteMsg{}`; closes on `Esc`
- [ ] Displays a filtered list combining: recent commands from the ring buffer (most recent first, max 20) and built-in actions (`Re-run last command`, `Copy last output`, `Fix with AI`)
- [ ] Fuzzy filtering updates the displayed list as the user types; built-in actions always appear below filtered results
- [ ] `up`/`down` or `k`/`j` navigate the list; `Enter` on a selected item emits the appropriate message and closes the palette
- [ ] `Esc` closes the palette without action, restoring focus to the input bar
- [ ] Palette renders as a centered overlay: bg=ink-1, border in ink-4 (rounded), title `"Command Palette"` in ink-8; width = `min(60, terminalWidth-4)`
- [ ] `go test ./internal/ui/palette/...` passes: fuzzy filter narrows results, `Esc` closes, built-in actions always present

**Tech notes**
- FR-19, FR-20 (PRD §6.5)
- Fuzzy matching: `github.com/sahilm/fuzzy` — add to `go.mod`
- The palette receives a `[]string` snapshot of recent commands at open time (not a live reference to the ring buffer)
- Built-in action items map to specific message types: re-run → `SubmitMsg`, copy → clipboard write, fix with AI → `f`-key handler
- DESIGN.md `--shadow-3` elevation and `--r-xl` radius — use a Lip Gloss `Border(lipgloss.RoundedBorder())` for the overlay

**Completion notes**
*(empty)*
