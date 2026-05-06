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
- [ ] If ACP `Connect` fails on startup, the root model sets `ShellOnly = true` and the header bar displays `"AI offline — add config: ~/.config/forge/config.toml"` in ink-6
- [ ] If `SendPrompt` returns an error mid-session, the affected AI card displays `"AI error — try again"` in crimson-500; no panic, no crash
- [ ] `f` key in shell-only mode shows the same offline hint in the header instead of calling ACP
- [ ] Killing the agent subprocess externally while Forge is running: `SendPrompt` returns `ErrNotConnected`; the root model recovers to shell-only mode without restarting
- [ ] `go test -race ./...` passes with a mock ACP client that returns `ErrNotConnected` on every call

**Tech notes**
- NFR-05 (PRD §7)
- `ShellOnly bool` is a field on the root model — set it to `true` on any unrecoverable ACP error
- The ACP client must detect subprocess death in a background goroutine (via `cmd.Wait()`) and update its internal connected state — `SendPrompt` then returns `ErrNotConnected` without blocking
- Inline hint display: append a status line to the header view, not a modal or popup — consistent with the "quiet by default" design principle (DESIGN.md §Principles)

**Completion notes**
*(empty)*
