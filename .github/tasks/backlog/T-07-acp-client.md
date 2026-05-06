---
id: T-07
status: backlog
agent: Code Expert
depends-on: [T-02, T-05]
branch: T-07/acp-client
---

## T-07 · ACP Client

**Context**
Implements the Forge side of the Agent Client Protocol over stdio transport. It manages the agent subprocess lifetime, sends prompts, and streams token responses back to the TUI. This is the sole integration point between Forge and any AI backend.

**Acceptance criteria**
- [ ] `internal/acp/client.go` exports `Client` with methods: `Connect(cfg *config.Config) error`, `SendPrompt(ctx context.Context, prompt string, onToken func(string)) error`, `Close() error`
- [ ] `Connect` starts the agent subprocess and sends `session/create` JSON-RPC 2.0 message over its stdin
- [ ] `SendPrompt` sends a `prompt/turn` request and calls `onToken` for each streamed token in the response
- [ ] `Close` sends `session/close` and waits for the subprocess to exit cleanly
- [ ] `SendPrompt` returns the sentinel `ErrNotConnected` if `Connect` was not called or the subprocess has exited — it never panics
- [ ] ACP framing: each JSON-RPC message is a single line terminated with `\n`
- [ ] `go test ./internal/acp/...` passes: mock subprocess test verifies `session/create` is sent on connect, `ErrNotConnected` is returned without a live process

**Tech notes**
- PRD §9: ACP methods for v1 are `session/create`, `prompt/turn`, `session/close`
- `var ErrNotConnected = errors.New("acp: not connected")` — sentinel error, do not use `fmt.Errorf` for this one
- `onToken` is called from the goroutine reading the subprocess stdout — callers must not block in `onToken`; they must use `tea.Send` or a channel
- The agent subprocess command comes from `cfg` — for tests, use a mock subprocess that reads JSON-RPC and echoes tokens
- NFR-05: `Connect` failure must return an error, not panic — the caller decides whether to degrade gracefully

**Completion notes**
*(empty)*
