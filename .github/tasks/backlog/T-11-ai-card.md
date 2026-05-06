---
id: T-11
status: backlog
agent: Code Expert
depends-on: [T-03, T-04]
branch: T-11/ai-card
---

## T-11 · AI Suggestion Card Component

**Context**
The inline AI card that appears inside a failed block after "Fix with AI" is triggered. It renders streaming tokens as they arrive and provides Run/Dismiss keyboard actions (FR-11 to FR-15).

**Acceptance criteria**
- [ ] `internal/ui/aicard/model.go` defines `Model` implementing `tea.Model`
- [ ] `View()` renders a card surface with bg=plasma-500 left accent border (1 cell wide), bg=ink-1 body, token text in plasma-300 (#c4b5ff)
- [ ] While `AICard.Streaming` is true, a blinking cursor `▋` is appended after the last token
- [ ] When streaming ends (`Streaming = false`), renders action hints at the bottom: `[Enter] Run  [Esc] Dismiss` in ink-6
- [ ] `Enter` emits `AcceptAIMsg { Command string }` (the last line of tokens is treated as the proposed command)
- [ ] `Esc` or `d` emits `DismissAIMsg{}`
- [ ] Key handling is only active when `AICard.Streaming` is false — cannot accept/dismiss while streaming
- [ ] `go test ./internal/ui/aicard/...` passes: token append, cursor present during stream, `AcceptAIMsg`/`DismissAIMsg` emitted on correct keys

**Tech notes**
- FR-13, FR-14, FR-15 (PRD §6.3)
- `Model` wraps a `block.AICard` value — it renders from data, does not own it
- Import `internal/theme` for plasma colors; import `internal/block` for `AICard` type
- DESIGN.md Plasma ramp: plasma-300 = `#c4b5ff`, plasma-500 = `#7c5cff`
- `AcceptAIMsg.Command` should be the last non-empty token line; the AI response format will clarify this in T-16

**Completion notes**
*(empty)*
