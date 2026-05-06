---
id: T-06
status: backlog
agent: Code Expert
depends-on: [T-01, T-02]
branch: T-06/shell-executor
---

## T-06 · Shell Executor

**Context**
Encapsulates `os/exec` command execution with real-time stdout/stderr streaming, exit code capture, wall-clock duration, and context-based cancellation. This is the engine behind FR-01 through FR-04.

**Acceptance criteria**
- [ ] `internal/exec/exec.go` exports `Run(ctx context.Context, dir string, command string, out chan<- string) (exitCode int, duration time.Duration, err error)`
- [ ] stdout and stderr are merged and sent line-by-line to `out` in real time as the process runs
- [ ] Cancelling `ctx` sends SIGTERM to the process; the function returns after the process exits (FR-04)
- [ ] `exitCode` is extracted from `*exec.ExitError`; a clean exit returns 0
- [ ] `duration` is wall-clock time from `cmd.Start()` to process termination
- [ ] The `out` channel is closed exactly once when the command terminates
- [ ] `command` is executed as `$SHELL -c "<command>"` to support pipes, aliases, and shell builtins
- [ ] `go test ./internal/exec/...` passes: exit-0 for `echo hello`, non-zero for `exit 1`, cancellation terminates process within 500ms

**Tech notes**
- FR-01 through FR-04 (PRD §6.1)
- Use `exec.CommandContext(ctx, shell, "-c", command)` where `shell = os.Getenv("SHELL")` with fallback `/bin/sh`
- Do NOT use `cmd.CombinedOutput()` — it buffers; instead use `cmd.StdoutPipe()` and `cmd.StderrPipe()` with `bufio.Scanner` in separate goroutines, merging into the `out` channel
- A `sync.WaitGroup` coordinates the two reader goroutines; close `out` only after both readers have exited
- Tests that launch real processes should use `//go:build !windows` build tag

**Completion notes**
*(empty)*
