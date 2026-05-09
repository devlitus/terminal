---
id: T-18
status: backlog
agent: Code Expert
depends-on: [T-14, T-15, T-16, T-17]
branch: T-18/root-assembly
---

## T-18 · Root TUI Assembly

**Context**
Composes all components into the root Bubble Tea model and produces the runnable `forge` binary. This is the final integration point where every subsystem is wired together and the layout is enforced.

**Acceptance criteria**
- [ ] `cmd/forge/main.go` initialises config (T-05), ACP client (T-07), session (T-15), and root model; calls `tea.NewProgram(root, tea.WithAltScreen()).Run()`
- [ ] Root model `View()` produces: header bar (1 line) + viewport (terminalHeight − 2 lines) + input bar (1 line)
- [ ] `tea.WindowSizeMsg` propagates correctly to header, viewport, and all child block models
- [ ] `q` key prompts `"Quit? (y/n)"` if any block is `StateRunning`; quits immediately if no commands are running
- [ ] `ctrl+c` at root level: if a block is running, cancels it (does not quit the app); if no block is running, quits
- [ ] Palette modal overlays the viewport when open — rendered last so it appears on top
- [ ] `go build ./cmd/forge` produces a working binary with zero errors
- [ ] End-to-end smoke test: launch binary, run `echo forge`, confirm block with success badge appears; run `exit 1`, confirm failed badge appears

**Tech notes**
- NFR-01 (first paint < 100ms), NFR-04 (80–220 column width support)
- All child models are embedded by value in the root model struct — avoid pointer fields to prevent nil-dereference panics
- Layout: `lipgloss.JoinVertical(lipgloss.Left, headerView, viewportView, inputView)` or simple `\n` concatenation
- Root model is the single owner of `*session.Session`, `*acp.Client`, and `*block.RingBuffer`
- `tea.WithAltScreen()` ensures the TUI takes the full terminal — required for correct layout
- Follow the message routing map in `ARCHITECTURE.md` exactly

**Completion notes**
*(empty)*
