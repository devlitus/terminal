---
id: T-19
status: backlog
agent: Code Expert
depends-on: [T-16, T-18]
branch: T-19/acp-degradation
---

## T-19 · Graceful ACP Degradation

**Context**
Ensures ACP failures — at startup or mid-session — never crash the TUI and that shell-only mode provides a clear, non-intrusive status hint. Directly addresses NFR-05.

**Acceptance criteria**
- [x] If ACP `Connect` fails on startup, the root model sets `ShellOnly = true` and the header bar displays `"AI offline — add config: ~/.config/forge/config.toml"` in ink-6
- [x] If `SendPrompt` returns an error mid-session, the affected AI card displays `"AI error — try again"` in crimson-500; no panic, no crash
- [x] `f` key in shell-only mode shows the same offline hint in the header instead of calling ACP
- [x] Killing the agent subprocess externally while Forge is running: `SendPrompt` returns `ErrNotConnected`; the root model recovers to shell-only mode without restarting
- [x] `go test -race ./...` passes with a mock ACP client that returns `ErrNotConnected` on every call

**Tech notes**
- NFR-05 (PRD §7)
- `ShellOnly bool` is a field on the root model — set it to `true` on any unrecoverable ACP error
- The ACP client must detect subprocess death in a background goroutine (via `cmd.Wait()`) and update its internal connected state — `SendPrompt` then returns `ErrNotConnected` without blocking
- Inline hint display: append a status line to the header view, not a modal or popup — consistent with the "quiet by default" design principle (DESIGN.md §Principles)

**Completion notes**
- `acp.Client` gained `mu sync.Mutex`, `waitDone chan struct{}`, `waitErr error`; background goroutine in `Connect()` calls `cmd.Wait()`, clears `connected` under mutex, closes `waitDone`; `Close()` waits on `waitDone` instead of calling `Wait()` a second time (double-Wait races on some platforms)
- `SendPrompt` reads `connected` under mutex — data-race-free against the death goroutine
- `header.Model` gained `statusHint string` + `SetStatusHint(*Model)` pointer-receiver method; `View()` appends the hint in ink-6 when non-empty
- `block.AICard` gained `ErrMsg string`; `blockui.Model.SetAIError` sets it; `aicard.View()` renders it in crimson-500
- `viewport.Model.SetAIError` delegates to `blockui.Model.SetAIError` and refreshes content
- `rootModel.shellOnly bool` set on `Connect` failure or `ACPDoneMsg.Err != nil`; `f` and `FixWithAIMsg` handlers check `shellOnly` first; `ACPDoneMsg` handler calls `vp.SetAIError` on error path
- `TestSendPromptAfterSubprocessDeath` exercises the concurrent death scenario with `-race`
