---
id: T-10
status: backlog
agent: Code Expert
depends-on: [T-03]
branch: T-10/input-bar
---

## T-10 · Input Bar Component

**Context**
The always-visible bottom input bar where users type shell commands and AI prompts. It must distinguish between shell commands and AI-prefixed input (`/` or `@agent`) before emitting a submit message.

**Acceptance criteria**
- [ ] `internal/ui/input/model.go` defines `Model` implementing `tea.Model`
- [ ] Uses `github.com/charmbracelet/bubbles/textinput` for text editing
- [ ] `View()` renders a single line: bg=ink-1, ember-500 `❯` prompt glyph, typed text in ink-9
- [ ] Pressing `Enter` emits `SubmitMsg { Input string, IsAIPrompt bool }` — `IsAIPrompt` is `true` when `Input` starts with `/` or `@`
- [ ] Input field is cleared after emit
- [ ] `ctrl+k` emits `OpenPaletteMsg{}` (palette handled elsewhere — this component only emits the message)
- [ ] Component does NOT execute commands or call any exec/ACP code
- [ ] `go test ./internal/ui/input/...` passes: `/fix it` yields `IsAIPrompt=true`, `git status` yields `IsAIPrompt=false`

**Tech notes**
- FR-12 (AI prompt detection), PRD §5.3 (input bar description)
- `textinput.Model` from `github.com/charmbracelet/bubbles/textinput`
- Do NOT intercept `ctrl+c` — that is handled at root model level for process kill (FR-04)
- Placeholder text: `"type a command or /prompt for AI…"` rendered in ink-6
- Height is always exactly 1 line

**Completion notes**
*(empty)*
