---
id: T-15
status: backlog
agent: Code Expert
depends-on: [T-09, T-13]
branch: T-15/session-mgmt
---

## T-15 · Session Management

**Context**
Implements the session-level commands: `clear` removes all blocks from the viewport, and `cd` updates the working directory so subsequent commands run in the correct location and the header reflects the change (FR-16 through FR-18).

**Acceptance criteria**
- [ ] `internal/session/session.go` tracks `Cwd string` and holds a reference to the shared `*block.RingBuffer`
- [ ] `clear` command (detected before exec) resets the ring buffer to empty and emits `ViewportClearMsg{}` to the root model
- [ ] `cd <path>` command (detected before exec, not passed to shell) resolves the new path, updates `Session.Cwd`, and emits `CwdChangedMsg { Dir: newPath }` to the header
- [ ] Relative `cd` paths are resolved with `filepath.Abs` relative to the current `Session.Cwd`
- [ ] `cd` without arguments is a no-op (do not change to `$HOME` — out of scope for v1)
- [ ] After `cd`, all subsequent shell commands receive `Dir = Session.Cwd` as their working directory
- [ ] On startup, `Session.Cwd` is initialised from `os.Getwd()`
- [ ] Integration test: `cd /tmp` → header updates; `clear` → viewport empty; `go test ./internal/session/...` passes

**Tech notes**
- FR-16, FR-17, FR-18 (PRD §6.4)
- Do NOT use `os.Chdir()` — it is global and affects the whole process; pass `Dir` to `exec.Cmd` per command
- `cd` detection: match against `^cd(\s+(.+))?$` before exec; `clear` is an exact string match
- `Session` is instantiated once in the root model and passed by pointer to command handlers
- `ViewportClearMsg` and `CwdChangedMsg` message types must match the names in ARCHITECTURE.md

**Completion notes**
*(empty)*
