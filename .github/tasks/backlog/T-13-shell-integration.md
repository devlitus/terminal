---
id: T-13
status: backlog
agent: Code Expert
depends-on: [T-06, T-10, T-12]
branch: T-13/shell-integration
---

## T-13 · Shell Execution Integration

**Context**
Wires the shell executor into the TUI: a command submitted from the input bar creates a block, streams output into it in real time, updates block state on exit, and handles ctrl+c kill. This is the first complete end-to-end user flow.

**Acceptance criteria**
- [ ] Root model handles `input.SubmitMsg` where `IsAIPrompt = false`
- [ ] On submit: creates a new `block.Block`, appends it to `RingBuffer`, appends a `blockui.Model` to the viewport, and launches exec via a `tea.Cmd` goroutine
- [ ] Streaming lines arrive as `ExecOutputMsg { BlockID uint64, Line string }` and are appended to the correct block's `Output` slice; viewport re-renders
- [ ] On exit, `ExecDoneMsg { BlockID uint64, ExitCode int, Duration time.Duration }` sets block state to `StateSuccess` or `StateFailed`
- [ ] `ctrl+c` while a block is `StateRunning` cancels that block's context; a "^C" line is appended to output; block transitions to `StateFailed`
- [ ] Each running command has its own `context.CancelFunc` stored in the root model keyed by `BlockID`
- [ ] Manual smoke test: type `echo forge`, press Enter — block appears with "forge" in output and success badge
- [ ] Manual smoke test: type `exit 1` — block shows failed badge

**Tech notes**
- FR-01 through FR-05 (PRD §6.1, §6.2)
- Message type names must match ARCHITECTURE.md exactly
- `cd` interception is handled in T-15 — this task does NOT need to implement it
- `tea.Cmd` for exec: return a `func() tea.Msg` that calls `exec.Run` and sends output lines via `tea.Send(p, ExecOutputMsg{...})`
- The `tea.Program` pointer `p` must be available to the goroutine — pass it via closure or use `tea.Send`

**Completion notes**
*(empty)*
