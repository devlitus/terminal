---
id: T-16
status: backlog
agent: Code Expert
depends-on: [T-07, T-11, T-13]
branch: T-16/acp-integration
---

## T-16 · ACP / AI Integration

**Context**
Wires the ACP client into the TUI: `f` on a failed block sends a "Fix with AI" prompt, `/`-prefixed input submits a free-form prompt, and streaming tokens update the AI card in real time (FR-11 through FR-15).

**Acceptance criteria**
- [ ] `f` key on a focused `StateFailed` block constructs a prompt from the block command and last 20 output lines, then calls `acp.Client.SendPrompt`
- [ ] `input.SubmitMsg` with `IsAIPrompt = true` sends the stripped text (without leading `/` or `@`) as a `prompt/turn`
- [ ] Streaming tokens arrive via the `onToken` callback and are dispatched as `ACPTokenMsg { BlockID uint64, Token string }` to the root model via `tea.Send`
- [ ] On `ACPTokenMsg`, the corresponding block's `AICard.Tokens` is appended and the AI card component re-renders
- [ ] When the ACP response completes, `AICard.Streaming` is set to `false`
- [ ] `AcceptAIMsg` from the AI card submits the suggested command through the shell execution flow (T-13)
- [ ] `DismissAIMsg` sets `AICard.Dismissed = true`; the card stops rendering
- [ ] If `config.IsShellOnly()` is true, `f` key shows a one-line inline hint instead of calling ACP
- [ ] Manual integration test: trigger "Fix with AI" on a failed block; confirm streaming tokens appear in the AI card

**Tech notes**
- FR-11 through FR-15 (PRD §6.3); NFR-05 (must not crash on ACP failure)
- Prompt template: `"Command: <cmd>\nError output:\n<last 20 lines>\nFix this command."`
- `onToken` is called from a goroutine — use `tea.Send(program, ACPTokenMsg{...})` to re-enter the Bubble Tea update loop safely
- Wrap all `acp.Client` calls in error checks; on error, set `ShellOnly = true` on the root model
- `f` is a no-op on `StateRunning` or `StateSuccess` blocks

**Completion notes**
*(empty)*
