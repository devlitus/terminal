---
id: T-09
status: backlog
agent: Code Expert
depends-on: [T-03]
branch: T-09/header-bar
---

## T-09 · Header Bar Component

**Context**
The persistent top bar that shows the current working directory and session metadata. It must update whenever a `cd` command changes the cwd (FR-17) and adapt to any terminal width.

**Acceptance criteria**
- [ ] `internal/ui/header/model.go` defines `Model` implementing `tea.Model`
- [ ] `View()` renders a single line: bg=ink-2, left-aligned cwd in ink-8, right-aligned block count in ink-6
- [ ] `CwdChangedMsg { Dir string }` received in `Update` updates the displayed path
- [ ] Long paths are truncated with a `…` prefix, preserving the trailing basename: e.g. `…/forge/api` — never truncate the final path component
- [ ] Width is received via `tea.WindowSizeMsg` and stored; `View()` uses it for correct truncation
- [ ] `go test ./internal/ui/header/...` passes: cwd display, truncation at short widths, `CwdChangedMsg` updates path

**Tech notes**
- FR-17 (PRD §6.4)
- Import `internal/theme` only; no business logic in this component
- `CwdChangedMsg` is defined here; ARCHITECTURE.md specifies whether it lives in a shared `msgs` package — follow that decision
- Height is always exactly 1 line — do not add padding or border
- Block count is formatted as `N blocks` in ink-6 on the right margin

**Completion notes**
*(empty)*
