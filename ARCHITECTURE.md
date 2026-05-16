# Forge — Architecture

> Version 1.0 · May 2026 · Authoritative reference for all implementation tasks.
> All message type names, package responsibilities, and design decisions in this document are canonical.
> Implementation tasks must not deviate from the names and contracts defined here without updating this document first.

---

## Table of Contents

1. [Overview](#1-overview)
2. [Package Responsibilities](#2-package-responsibilities)
3. [Bubble Tea Message Types](#3-bubble-tea-message-types)
4. [Data Flow: Shell Execution](#4-data-flow-shell-execution)
5. [Data Flow: ACP / AI Integration](#5-data-flow-acp--ai-integration)
6. [Concurrency Model](#6-concurrency-model)
7. [Layout Zones](#7-layout-zones)
8. [Trade-offs / ADRs](#8-trade-offs--adrs)
9. [Agent System (Agentic Loop + Tool Calls)](#9-agent-system-agentic-loop--tool-calls)

---

## 1. Overview

Forge is a TUI terminal application built on Bubble Tea's Elm-style architecture. Every shell command produces a discrete, addressable **Command Block** — a self-contained unit of header, input, and streamed output — rendered in a scrollable viewport between a persistent header bar and input bar. An HTTP client (`internal/acp`) connects Forge directly to any OpenAI-compatible endpoint, enabling inline AI assistance and a stateful agentic loop with tool calls.

---

## 2. Package Responsibilities

| Package | Import Path | Responsibility |
|---------|-------------|---------------|
| `main` | `github.com/forge-tui/forge/cmd/forge` | Program entry point: loads config, constructs `acp.Client` + `agent.Agent`, builds the root Bubble Tea model (`rootModel`), and calls `tea.NewProgram().Run()`. Also owns block creation — `SubmitMsg` is handled here, which adds to the ring buffer and calls `forgeExec.Start()`. |
| `config` | `github.com/forge-tui/forge/internal/config` | Loads and validates `~/.config/forge/config.toml`; exposes a typed `Config` struct; `IsShellOnly()` returns true when `api_base` is empty |
| `theme` | `github.com/forge-tui/forge/internal/theme` | Declares all Lip Gloss `lipgloss.Color` constants mapped from the design token table; the single source of truth for every color value in the TUI |
| `block` | `github.com/forge-tui/forge/internal/block` | Defines the `Block` data model (ID, command, cwd, state, exit code, duration, output lines, `AICard`) and `RingBuffer` (cap 500, front-drop eviction) |
| `exec` | `github.com/forge-tui/forge/internal/exec` | Starts shell commands via `os/exec` + `io.Pipe`; owns the per-command goroutine that reads stdout/stderr and emits `ExecOutputMsg` and `ExecDoneMsg` via `p.Send()` |
| `acp` | `github.com/forge-tui/forge/internal/acp` | HTTP client for any OpenAI-compatible endpoint. `SendPrompt` does single-turn streaming; `Chat` supports multi-turn with tool calls (SSE, OpenAI wire format). No subprocess — no stdio. |
| `agent` | `github.com/forge-tui/forge/internal/agent` | Agentic loop on top of `acp`: maintains conversation history, dispatches tool calls (`run_command`, `read_file`, `list_dir`, `get_cwd`), confirms destructive commands via `AgentConfirmMsg`, loops until the model replies with no tool calls |
| `session` | `github.com/forge-tui/forge/internal/session` | Tracks current working directory; intercepts built-ins (`cd`, `clear`) before they reach the shell; dispatches `CwdChangedMsg` on `cd` and `ViewportClearMsg` on `clear` |
| `ui/block` | `github.com/forge-tui/forge/internal/ui/block` | Bubble Tea component for a single Command Block; renders header, command line, scrollable output, and AI card; exposes `SetFocused`, `AppendOutput`, `AppendAIToken`, `SetDone`, `SetAIError`, `DismissAICard` |
| `ui/header` | `github.com/forge-tui/forge/internal/ui/header` | Renders the 1-line top bar: CWD and optional status hint (e.g. "AI offline…"); updates on `CwdChangedMsg`; shows a second line when `SetStatusHint` is non-empty |
| `ui/input` | `github.com/forge-tui/forge/internal/ui/input` | Renders the 1-line bottom input bar. Two modes: chat (`⬡`, plasma-500) and shell (`❯`, ember-500), toggled by `!`. Emits `SubmitMsg` on Enter and `OpenPaletteMsg` on `ctrl+k` |
| `ui/aicard` | `github.com/forge-tui/forge/internal/ui/aicard` | Renders the AI Suggestion Card embedded in a block; streams `ACPTokenMsg` tokens; shows Run/Dismiss actions after `ACPDoneMsg`; handles Enter (accept) and Esc/`ctrl+d` (dismiss) |
| `ui/viewport` | `github.com/forge-tui/forge/internal/ui/viewport` | Manages the scrollable list of `ui/block` components; routes messages to the correct child by ID; handles `up`/`down` block navigation, `ctrl+y` copy, `ctrl+r` re-run, and AI card key events |
| `ui/palette` | `github.com/forge-tui/forge/internal/ui/palette` | Renders the command palette overlay triggered by `ctrl+k`; implements fuzzy filtering over the 20 most-recent commands plus built-in actions (Re-run, Copy, Fix with AI); emits the selected action on `Enter` |

---

## 3. Bubble Tea Message Types

All message types are defined in a single file: `internal/messages/messages.go` (package `messages`). This package has **no dependencies on any other internal package** — it holds only plain structs so that every other package can import it without creating import cycles.

Implementation tasks must use these exact type names. No aliases, no local redefinitions.

---

### 3.1 `ExecOutputMsg`

Carries one chunk of stdout/stderr bytes from a running shell command.

```go
// ExecOutputMsg is emitted by the exec goroutine each time the pipe
// yields a non-empty read from a running command's stdout/stderr.
type ExecOutputMsg struct {
    BlockID string // identifies the target block
    Data    []byte // raw bytes; may be partial lines
}
```

**Sent by:** `internal/exec` — the per-command goroutine, via a `tea.Cmd` that blocks on `io.Pipe` reads.  
**Consumed by:** `internal/ui/viewport` — routes to the child `ui/block` component matching `BlockID`.

---

### 3.2 `ExecDoneMsg`

Signals that a shell command has exited and all output has been flushed.

```go
// ExecDoneMsg is emitted by the exec goroutine when the command exits
// and the output pipe is fully drained.
type ExecDoneMsg struct {
    BlockID  string
    ExitCode int           // 0 = success, non-zero = failure
    Duration time.Duration // wall-clock elapsed time
}
```

**Sent by:** `internal/exec` — emitted once per command, after the pipe EOF.  
**Consumed by:** `internal/ui/viewport` (updates block state and badge) and `internal/session` (updates the data model).

---

### 3.3 `ACPTokenMsg`

Carries a single streamed token from the ACP agent response.

```go
// ACPTokenMsg is emitted by the ACP streaming goroutine for each
// token received from the agent. Consumers must append, not replace.
type ACPTokenMsg struct {
    BlockID string // identifies which AI card receives this token
    Token   string // a single text fragment; may be a partial word
}
```

**Sent by:** `internal/agent` — one message per SSE content delta from the HTTP streaming response. Tool call notifications are also sent as `ACPTokenMsg` with an inline `⚙ tool_name…` prefix.  
**Consumed by:** `internal/ui/viewport` → `internal/ui/block` (which holds the AI card) for the matching `BlockID`.

---

### 3.4 `ACPDoneMsg`

Signals the end of an ACP agent response stream.

```go
// ACPDoneMsg is emitted by the ACP streaming goroutine when the
// agent signals end-of-stream or when an error terminates the call.
type ACPDoneMsg struct {
    BlockID string
    Err     error // nil on clean completion; non-nil on transport or agent error
}
```

**Sent by:** `internal/agent` — emitted once per `agent.Send()` call when the agentic loop ends (either clean completion or error).  
**Consumed by:** `internal/ui/viewport` → `internal/ui/block` — transitions the AI card from streaming state to Run/Dismiss action state. When `Err != nil`, the root model sets `shellOnly = true` and the card renders an error message (NFR-05).

---

### 3.5 `BlockFocusedMsg`

Signals that keyboard focus has moved to a specific block.

```go
// BlockFocusedMsg is emitted by the viewport when the user navigates
// focus with arrow keys. An empty BlockID means no block is focused.
type BlockFocusedMsg struct {
    BlockID string
}
```

**Sent by:** `internal/ui/viewport` — on `up` / `down` arrow key navigation.  
**Consumed by:** `internal/ui/viewport` itself — `focusedIdx` is updated and `renderBlocks()` re-renders all blocks with the correct focus state (ember-500 border on the focused block).

---

### 3.6 `CwdChangedMsg`

Signals that the working directory has changed after a successful `cd` command.

```go
// CwdChangedMsg is emitted by session.Handle() when a cd command
// completes and the working directory has changed.
type CwdChangedMsg struct {
    Cwd string // absolute path of the new working directory
}
```

**Sent by:** `internal/session` — after resolving and validating the new path.  
**Consumed by:** `internal/ui/header` (updates the cwd display) and `internal/ui/input` (used for prompt rendering).

---

### 3.7 `OpenPaletteMsg`

Signals that `ctrl+k` was pressed and the command palette should open.

```go
// OpenPaletteMsg is emitted by the input bar when the user presses
// ctrl+k. The root model switches rendering to the palette overlay.
type OpenPaletteMsg struct{}
```

**Sent by:** `internal/ui/input` — on `ctrl+k` key event.  
**Consumed by:** `cmd/forge` root model — sets a `paletteOpen bool` flag that causes `View()` to render `ui/palette` as an overlay instead of the normal layout.

---

### 3.8 `ViewportClearMsg`

Signals the `clear` built-in command: remove all blocks from the viewport.

```go
// ViewportClearMsg is emitted when the user runs the `clear` built-in.
// The viewport drops all rendered blocks; the session ring buffer is also reset.
type ViewportClearMsg struct{}
```

**Sent by:** `internal/ui/input` — when the entered command is exactly `"clear"`.  
**Consumed by:** `internal/ui/viewport` (clears all child block components) and `internal/session` (resets the block slice to empty).

---

## 4. Data Flow: Shell Execution

```
User types command, presses Enter
        │
        ▼
  ui/input.Update(tea.KeyMsg{"enter"})
        │  emits SubmitMsg{Input: cmd}
        ▼
  rootModel.Update(SubmitMsg)  [cmd/forge/main.go]
        │  1. session.Handle(cmd) — intercepts built-ins (cd, clear); returns if handled
        │  2. b := block.Block{Command: cmd, Dir: cwd, State: Running, StartedAt: now}
        │  3. b = rb.Add(b)          ← ring buffer (cap 500, front-drop)
        │  4. vp.AppendBlock(b)      ← viewport adds a new ui/block child directly
        │  5. returns forgeExec.Start(program, blockID, cmd, dir, ctx)
        ▼
  exec.Start() → launches goroutine
        │
        │  goroutine: reads io.Pipe (stdout+stderr merged)
        │  communicates via program.Send() — never writes to model state
        │
        ├── chunk available ──► program.Send(ExecOutputMsg{BlockID, Data})
        │                              │
        │                              ▼
        │                       rootModel.Update(ExecOutputMsg)
        │                              │ delegates to vp.Update(msg)
        │                              ▼
        │                       viewport routes by BlockID → ui/block.AppendOutput()
        │                              └─► tea re-renders the block
        │
        └── pipe EOF ──────────► program.Send(ExecDoneMsg{BlockID, ExitCode, Duration})
                                        │
                                        ▼
                                 rootModel.Update(ExecDoneMsg)
                                        │ cancels the context for that blockID
                                        │ delegates to vp.Update(msg)
                                        ▼
                                 viewport routes by BlockID → ui/block.SetDone()
                                        └─► sets State = Success | Failed, badge re-renders
```

**Key invariant:** The goroutine in `exec.Start` never writes to shared memory. It communicates exclusively via `program.Send()`, which enqueues messages into the Bubble Tea runtime's internal channel. The runtime delivers them to `Update` on the main goroutine — no locks are needed in any `Update` or `View` function.

---

## 5. Data Flow: ACP / AI Integration

```
User presses ctrl+f on a focused failed block (or selects "Fix with AI" from palette)
        │
        ▼
  rootModel.Update(tea.KeyMsg{"ctrl+f"})  [cmd/forge/main.go]
        │  1. checks shellOnly — if true, shows offline hint, returns
        │  2. looks up focused block; checks State == StateFailed
        │  3. builds prompt: "Command: X\nError output:\n...\nFix this command."
        │  4. vp.InitAICard(blockID)   ← shows streaming placeholder in the block
        │  5. returns agent.Send(program, blockID, prompt)
        ▼
  agent.Send() → goroutine starts (agentic loop)
        │  Appends system prompt + user message to history
        │
        ├─ LOOP ─────────────────────────────────────────────────────────────────┐
        │    │                                                                    │
        │    ├─ client.Chat(ctx, history, toolDefs, onToken)                     │
        │    │       onToken → program.Send(ACPTokenMsg{BlockID, token})         │
        │    │                                                                    │
        │    ├─ if toolCalls == nil:                                              │
        │    │       append assistant reply to history                            │
        │    │       program.Send(ACPDoneMsg{BlockID}) ← LOOP EXIT               │
        │    │                                                                    │
        │    └─ else (model requested tools):                                     │
        │            append assistant tool-call turn to history                  │
        │            for each tool call:                                         │
        │              program.Send(ACPTokenMsg{"\n⚙ tool_name…\n"})            │
        │              if run_command → program.Send(AgentConfirmMsg)            │
        │                              block goroutine on reply channel          │
        │                              rootModel.Update(AgentConfirmMsg)         │
        │                              → shows "Run: X  (y/n)" in input bar     │
        │                              user presses y/n → reply sent             │
        │              result = tool.Execute(ctx, args)                          │
        │              append tool result to history                             │
        │            continue LOOP ───────────────────────────────────────────────┘
        │
        ├── ACPTokenMsg received ──► rootModel.Update → vp.Update → block.AppendAIToken
        │                                └─► tea re-renders block (streaming text)
        │
        └── ACPDoneMsg received
              │  Err == nil → vp.Update → block.SetAIStreamDone
              │               AI card shows [Enter=Run] [ctrl+d=Dismiss] actions
              │  Err != nil → root sets shellOnly=true
              │               vp.SetAIError(blockID, "AI error — try again")
              └─► tea re-renders

  User presses Enter on the AI card
        │
        ▼
  viewport.Update(tea.KeyMsg{"enter"}) → block.Update(enter)
        │  emits AcceptAIMsg{Command: suggestedCmd}
        ▼
  rootModel.Update(AcceptAIMsg)
        │  dismisses card → emits DismissAIMsg
        └─► creates new block, calls forgeExec.Start(...)  ← identical to §4

  User presses Esc or ctrl+d on [Dismiss]
        │
        ▼
  viewport.Update → block.Update → emits DismissAIMsg
        └─► vp.Update(DismissAIMsg) → block.DismissAICard(), re-renders
```

**Graceful degradation (NFR-05):** If `api_base` is empty or the HTTP request fails, `ACPDoneMsg.Err` is non-nil. The root model sets `shellOnly = true`, the AI card renders `"AI error — try again"` in crimson-500, and the rest of the TUI continues operating normally.

---

## 6. Concurrency Model

### Goroutines

| Goroutine | Owner | Lifetime | Communicates via |
|-----------|-------|----------|-----------------|
| Bubble Tea event loop | `tea.Program` | Process lifetime | `tea.Msg` channel (internal) |
| Per-command exec reader | `internal/exec` | Duration of one shell command | `program.Send(ExecOutputMsg)` / `program.Send(ExecDoneMsg)` |
| Agent agentic loop | `internal/agent` | Duration of one `agent.Send()` call | `program.Send(ACPTokenMsg)` / `program.Send(ACPDoneMsg)` / `program.Send(AgentConfirmMsg)` |

At peak load there is at most one exec goroutine per running block plus one agent goroutine. Multiple commands can be started in parallel (each gets its own context + cancel); the ring buffer and viewport are updated synchronously in `Update`.

### Safety invariant

Bubble Tea's `Update(msg)` function is called **sequentially on the main goroutine**. Background goroutines (exec reader, ACP reader) never write to any field on any model or component. They only communicate by having their `tea.Cmd` closure send a message into the Bubble Tea runtime's internal channel. The runtime picks up the message and delivers it to `Update` on the main goroutine.

This means:
- No mutexes are needed in any `Update` or `View` function.
- No `sync/atomic` is needed on model fields.
- Data races are structurally impossible within the Bubble Tea model tree.

The only shared resource is the `io.Pipe` between `exec.Start` (writer: `os/exec` internals) and the reader goroutine — that sharing is safe because `io.Pipe` is explicitly designed for concurrent use.

### tea.Cmd composition

Background work is initiated by returning `tea.Cmd` values from `Update`. Multiple concurrent commands are composed with `tea.Batch`. The exec reader loop is structured as a recursive `tea.Cmd`:

```go
// Conceptual structure — not a literal implementation
func readNextChunk(r io.Reader, blockID string) tea.Cmd {
    return func() tea.Msg {
        buf := make([]byte, 4096)
        n, err := r.Read(buf)
        if n > 0 {
            return ExecOutputMsg{BlockID: blockID, Data: buf[:n]}
        }
        if err != nil { // io.EOF or real error
            return ExecDoneMsg{BlockID: blockID, ExitCode: ..., Duration: ...}
        }
        return nil
    }
}
```

`Update` handles `ExecOutputMsg` and re-issues `readNextChunk` as the next `tea.Cmd`, creating a pull loop driven by the Bubble Tea runtime rather than a tight goroutine spin.

---

## 7. Layout Zones

The terminal window is divided into three fixed zones. Heights are computed from `tea.WindowSizeMsg` and stored on the root model. Every resize event reflows all three zones.

```
┌───────────────────────────────────────────────────── termWidth ──┐
│  Header Bar                                              1 line   │  bg: ink-2 (#161922)
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Block Viewport                              termHeight − 2 lines │  bg: ink-0 (#0a0b10)
│  (scrollable list of Command Blocks)                              │
│                                                                   │
├──────────────────────────────────────────────────────────────────┤
│  Input Bar                                               1 line   │  bg: ink-1 (#0f1118)
└──────────────────────────────────────────────────────────────────┘
```

### Zone: Header Bar (1–2 lines, top)

**Height:** 1 terminal row normally; expands to 2 rows when `SetStatusHint` is non-empty (e.g. "AI offline — add config: ~/.config/forge/config.toml").  
**Content:** `Forge` · `cwd` · optional status hint on a second line (ink-6).

```go
// Lip Gloss style (internal/ui/header)
headerStyle := lipgloss.NewStyle().
    Background(lipgloss.Color("#161922")). // ink-2
    Foreground(lipgloss.Color("#8b91a8")). // ink-7
    Width(termWidth).
    Padding(0, 1)

cwdStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("#c5c9d6")). // ink-8
    Bold(true)
```

### Zone: Block Viewport (fills remaining height)

**Height:** `termHeight - 2` rows.  
**Behaviour:** Vertically scrollable. Renders a Lip Gloss-joined list of `ui/block.View()` strings. Auto-scrolls to the bottom when a new block is appended or output is streaming. Stops auto-scroll when the user navigates focus upward.

```go
// Lip Gloss style (internal/ui/viewport)
viewportStyle := lipgloss.NewStyle().
    Background(lipgloss.Color("#0a0b10")). // ink-0
    Width(termWidth).
    Height(termHeight - 2)

// Focused block border (internal/ui/block)
focusedBorderStyle := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("#ff6b3d")). // ember-500
    Padding(0, 1)

unfocusedBorderStyle := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("#2a2f40")). // ink-4
    Padding(0, 1)
```

**Block header colours:**

```go
// Status badge colours (internal/theme)
const (
    ColorSuccess = lipgloss.Color("#22c97e") // mint-500
    ColorFailed  = lipgloss.Color("#ef4444") // crimson-500
    ColorRunning = lipgloss.Color("#eab308") // solar-500
    ColorMuted   = lipgloss.Color("#5a6178") // ink-6
    ColorBody    = lipgloss.Color("#c5c9d6") // ink-8
    ColorAI      = lipgloss.Color("#7c5cff") // plasma-500
    ColorAccent  = lipgloss.Color("#ff6b3d") // ember-500
)
```

### Zone: Input Bar (1 line, bottom)

**Height:** always 1 terminal row.  
**Content:** mode-dependent prompt glyph + text field. Two modes:

| Mode | Glyph | Prompt color | Trigger |
|------|-------|-------------|---------|
| Chat (default) | `⬡` | plasma-500 (`#7c5cff`) | startup; `!` toggle |
| Shell | `❯` | ember-500 (`#ff6b3d`) | `!` toggle; `!command` one-shot |

When a quit prompt or agent confirmation is pending, the input bar is replaced by a temporary status line rendered in solar-500.

```go
// Lip Gloss style (internal/ui/input)
inputBarStyle := lipgloss.NewStyle().
    Background(lipgloss.Color("#0f1118")). // ink-1
    Width(termWidth).
    Padding(0, 1)
```

**Width responsiveness (NFR-04):** All zone widths are set from `tea.WindowSizeMsg.Width`. Supported range is 80–220 columns. The block output area clips long lines rather than wrapping to preserve command-output fidelity.

---

## 8. Trade-offs / ADRs

### ADR-01: Bubble Tea (Elm Architecture) over direct termbox/tcell

**Status:** Accepted

**Context:**  
Forge needs concurrent output streaming (multiple goroutines writing to the screen), keyboard navigation, and composable UI components (blocks, palette, AI card). Implementing this directly with a low-level terminal library (tcell, termbox-go) would require hand-written event loops, manual synchronisation between goroutines and render calls, and bespoke component abstractions.

**Decision:**  
Use Bubble Tea v1 (`github.com/charmbracelet/bubbletea` v1.x) as the TUI framework.

**Rationale:**

| Concern | Bubble Tea answer |
|---------|------------------|
| Concurrent goroutines writing to UI | `tea.Cmd` channels serialise all updates onto one goroutine — no locks needed |
| Composable components | `tea.Model` interface is recursive; viewport, block, aicard, palette are all models |
| Scrollable output | Bubbles `viewport` component or manual Lip Gloss joins with offset tracking |
| Keyboard event routing | `tea.KeyMsg` dispatched by the runtime; each component handles its own keys |
| Testing | `Update(msg)` is a pure function — components are unit-testable without a real terminal |

**Consequences:**
- (+) No shared-state bugs in the render path.
- (+) Components are independently testable.
- (+) Lip Gloss integrates natively for styling.
- (−) The Elm-style message-passing model has a learning curve for contributors unfamiliar with it.
- (−) Deep message routing (root → viewport → block → aicard) requires discipline; the `messages` package (§3) eliminates import cycles.
- (−) Bubble Tea v2 changed its module path to `charm.land/bubbletea/v2` and has breaking API changes; this project pins v1 to remain compatible with the existing Bubbles and Lip Gloss ecosystem.

---

### ADR-02: Direct HTTP to OpenAI-compatible endpoint (no subprocess)

**Status:** Accepted

**Context:**  
Two options were considered for connecting Forge to an LLM: (a) spawn a local `forge-agent` subprocess and communicate over ACP/stdio, or (b) call the OpenAI-compatible HTTP API directly from `internal/acp`.

**Decision:**  
Call the OpenAI-compatible `/chat/completions` endpoint directly via HTTP + SSE. No subprocess is spawned.

**Rationale:**

| Concern | Subprocess (ACP/stdio) | Direct HTTP |
|---------|----------------------|-------------|
| Dependency | Requires `forge-agent` binary on `$PATH` | Only needs an API URL in config |
| Startup | Must resolve binary, spawn process | Instant — just an HTTP client |
| Subprocess death handling | Background goroutine + `cmd.Wait()` | No subprocess to monitor |
| Debugging | Must intercept pipe (`tee`) | Inspectable with curl or any HTTP proxy |
| Model flexibility | Agent binary controls model choice | `model` field in config; works with Ollama, OpenRouter, DeepSeek, etc. |
| Tool call support | Depends on agent implementation | Native via OpenAI `tools` field |

**Consequences:**
- (+) No binary dependency; zero setup beyond a config file.
- (+) Works with any OpenAI-compatible endpoint (Ollama, DeepSeek, OpenRouter, etc.) by changing `api_base`.
- (+) Tool calls are first-class via the `tools` field — no custom protocol extension needed.
- (−) Forge makes HTTP calls directly, which means it needs the API key in config (vs. an agent subprocess acting as a credential proxy).
- (−) Debugging requires HTTP inspection tooling rather than simply reading stdio output.

---

### ADR-03: Ring Buffer Cap of N=500 Blocks (resolves OQ-2)

**Status:** Accepted

**Context:**  
PRD OQ-2 asks: should the block viewport hold a maximum N blocks in memory, and if so, what is N?  
PRD §13 requires "no crash on 500+ blocks in a session." Without a cap, a very long session accumulates unbounded output in memory, violating NFR-03 (no memory leaks over 1 hour).

**Decision:**  
Cap the block list in `internal/session` at **N = 500 blocks**. When the 501st block is added, the oldest block (index 0) is evicted from the slice. The `ui/viewport` renders only the blocks currently in the session's list.

**Memory bound rationale:**

| Scenario | Avg output / block | 500 blocks total |
|----------|--------------------|-----------------|
| Typical CLI (git, npm) | ~5 KB | ~2.5 MB |
| Verbose builds (go build) | ~50 KB | ~25 MB |
| Pathological (long test run) | ~100 KB | ~50 MB |

The 50 MB ceiling at the pathological case is acceptable for a developer TUI on a modern laptop. If a single command produces output in excess of 100 KB, the block's output buffer itself should be capped separately (implementation detail of `ui/block`; suggested limit: 1 MB per block, truncating the oldest lines within a block using a line ring buffer).

**Eviction policy:**  
Simple front-drop: `blocks = blocks[1:]`. No LRU, no priority. The user sees the most recent 500 commands, which matches the mental model of a terminal session.

**Implementation:**

```go
// internal/session — conceptual structure
const MaxBlocks = 500

func (s *Session) appendBlock(b block.Block) {
    s.blocks = append(s.blocks, b)
    if len(s.blocks) > MaxBlocks {
        s.blocks = s.blocks[len(s.blocks)-MaxBlocks:]
    }
}
```

**Impact on FR-06:**  
FR-06 states "blocks scroll vertically; the viewport shows the most recent N blocks." The N in FR-06 is satisfied by this cap. Blocks older than 500 are not visible, which is consistent with PRD §13's success criterion (no crash at 500+ blocks) — the criterion requires survivability, not infinite retention.

**Impact on `clear` (FR-18):**  
`ViewportClearMsg` resets `s.blocks = s.blocks[:0]`, restoring full capacity. The cap is a session-lifetime watermark, not a fixed window.

**Consequences:**
- (+) Hard memory ceiling of ~50 MB for command output regardless of session length.
- (+) Simple implementation — no complex data structure, just a slice with a trim.
- (+) Satisfies NFR-03 (no memory leaks over 1 hour).
- (−) Commands from more than 500 executions ago are lost. Mitigated by the fact that shell history is handled by the user's shell independently of Forge.
- (−) Users running very short commands rapidly (e.g., benchmarking scripts) will see older blocks disappear. This is the expected trade-off and is documented behaviour.

---

## 9. Agent System (Agentic Loop + Tool Calls)

> **Status:** Implemented · May 2026

### 9.1 Problem Statement

The current `acp.Client.SendPrompt` sends a single user message with no history, no system prompt, and no tool definitions. The model cannot take actions, only generate text. To be useful in a terminal context the agent must be able to:

1. Maintain conversation context across turns (multi-turn history).
2. Invoke tools (shell commands, file reads, directory listings) autonomously.
3. Continue reasoning after each tool result until it produces a final answer (the agentic loop).

### 9.2 Package Structure

No existing package is rewritten. The changes are purely additive, except for two small extensions to existing files.

```
internal/
  acp/
    client.go    — EXTENDED: Chat() added alongside existing SendPrompt
    types.go     — NEW: OpenAI-compatible wire types (Message, ToolCall, ToolDef, …)
  agent/
    agent.go     — NEW: Agent struct, Send(), Reset()
    tools.go     — NEW: Tool type + 4 MVP tool implementations
    prompt.go    — NEW: systemPrompt constant
  messages/
    messages.go  — EXTENDED: AgentConfirmMsg added
```

**Why a new `internal/agent/` package instead of extending `internal/acp/`?**

`acp` is a transport layer: it speaks HTTP to an OpenAI-compatible endpoint. Conversation history, system prompts, tool dispatch, and the retry loop are orchestration concerns — they belong one layer above the transport. Mixing them into `acp` would make that package own two very different responsibilities, violating the single-responsibility principle and making the existing tests harder to maintain.

`agent` imports `acp`; `acp` does not import `agent`. No circular dependency.

### 9.3 Wire Types (`internal/acp/types.go`)

These are the OpenAI API JSON shapes. They belong in `acp` because `acp.Client.Chat()` uses them directly in the HTTP request/response, and `agent` imports `acp` to reference them. Defining them here avoids a shared "types" package while keeping `acp` free from agent logic.

```go
// Message is one entry in the OpenAI messages array.
// Role is one of: "system" | "user" | "assistant" | "tool".
type Message struct {
    Role       string     `json:"role"`
    Content    string     `json:"content,omitempty"`
    ToolCalls  []ToolCall `json:"tool_calls,omitempty"`   // set by assistant turns
    ToolCallID string     `json:"tool_call_id,omitempty"` // set on role="tool" turns
    Name       string     `json:"name,omitempty"`         // tool name on role="tool"
}

// ToolCall is a single function invocation requested by the model.
type ToolCall struct {
    ID       string           `json:"id"`
    Type     string           `json:"type"` // always "function"
    Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
    Name      string `json:"name"`
    Arguments string `json:"arguments"` // JSON-encoded object; parsed by the tool
}

// ToolDef is the schema sent to the model to advertise an available function.
type ToolDef struct {
    Type     string      `json:"type"` // always "function"
    Function FunctionDef `json:"function"`
}

type FunctionDef struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    Parameters  json.RawMessage `json:"parameters"` // JSON Schema object
}

// ChatRequest is the full body of a POST /chat/completions call.
type ChatRequest struct {
    Model    string    `json:"model"`
    Messages []Message `json:"messages"`
    Tools    []ToolDef `json:"tools,omitempty"`
    Stream   bool      `json:"stream"`
}

// ChatResponse is the assembled result of one Chat() call.
// Content holds streamed text (may be empty if the model only produced tool calls).
// ToolCalls holds any function invocations the model requested.
type ChatResponse struct {
    Content   string
    ToolCalls []ToolCall
}
```

### 9.4 `acp.Client` Extension

One new method is added to `acp.Client`. `SendPrompt` is preserved unchanged for backward compatibility with existing tests.

```go
```go
// Chat sends a multi-turn conversation to POST /chat/completions.
// onToken is called for each streamed content delta; tool call deltas are
// accumulated internally. Returns the assembled tool calls (nil if the model
// replied with content only). Chat blocks until the SSE stream ends.
func (c *Client) Chat(ctx context.Context, msgs []Message, tools []ToolDef, onToken func(string)) ([]ToolCall, error)

// Model returns the configured model name.
func (c *Client) Model() string
```

`Chat` accumulates tool call deltas by index across SSE chunks and returns them as `[]ToolCall` when the stream ends. Content tokens are forwarded via `onToken` as they arrive.

### 9.5 Tool Layer (`internal/agent/tools.go`)

```go
// Tool pairs an OpenAI ToolDef (the schema sent to the model) with a Go
// function that executes when the model requests this tool.
type Tool struct {
    Def     acp.ToolDef
    Execute func(ctx context.Context, args json.RawMessage) (string, error)
}
```

**MVP tool set — exactly four tools:**

| Tool name | Description | Uses |
|-----------|-------------|------|
| `run_command` | Execute a shell command in the current working directory; return combined stdout+stderr | `internal/exec.Run` |
| `read_file` | Read a file's contents given an absolute or CWD-relative path | `os.ReadFile` + path sanitisation |
| `list_dir` | List files and directories at a path (names + is-dir flag) | `os.ReadDir` |
| `get_cwd` | Return the current working directory (no arguments) | `cwd()` getter |

**Why these four?** They cover the essential terminal-AI use cases (run → observe → fix loop) without adding surface area that is hard to audit. Tool calls run as the user's OS process — there is no sandbox. Keeping the set minimal reduces the blast radius of a misbehaving model.

**Security note on `run_command`:** The tool executes whatever command string the model produces. There is no allowlist. This is an explicit trust-in-the-local-model decision (the user controls Ollama). A follow-up hardening task should add a confirmation step for patterns matching `rm -rf`, `git push --force`, `sudo`, and pipe-to-shell idioms (`curl … | sh`). This is documented as a known risk, not overlooked.

**`read_file` path handling:** The path is normalised with `filepath.Clean` and resolved relative to CWD if not absolute. There is no restriction preventing traversal (e.g., `../../etc/passwd`) because the local model context makes this a non-threat for the v1 use case. If Forge ever gains remote-agent support this must be revisited.

### 9.6 Agent (`internal/agent/agent.go`)

```go
// Agent owns conversation history for one session and runs the agentic loop.
// It is safe for concurrent use; the loop goroutine and potential future
// callers serialise via mu.
type Agent struct {
    history []acp.Message   // grows each turn; never persisted to disk
    tools   []Tool
    client  *acp.Client
    cwd     func() string   // live getter — returns session.Session.Cwd at call time
    mu      sync.Mutex
}

```go
// New constructs an Agent with the four MVP tools pre-registered.
// confirmFn is called by run_command before executing; returning false cancels it.
func New(client *acp.Client, cwd func() string, confirmFn func(command string) bool) *Agent

// Send starts the agentic loop for userInput as a tea.Cmd.
// The returned cmd runs in a goroutine. It emits intermediate messages via
// p.Send() and returns ACPDoneMsg as its final tea.Msg when the loop ends.
func (a *Agent) Send(p *tea.Program, blockID, userInput string) tea.Cmd

// Reset clears conversation history. Called when the user runs `clear`.
func (a *Agent) Reset()
```

`history` is a flat `[]acp.Message`. A system prompt is prepended to history on the first turn and stored there. Tool call notifications are sent as `ACPTokenMsg` with an inline `⚙ tool_name…` prefix rather than as a dedicated message type.

### 9.7 System Prompt (`internal/agent/prompt.go`)

The system prompt is a package-level constant `systemPrompt` (not a function). It is appended to history on the first turn and does not change on `cd`.

```
You are Forge, an AI assistant embedded in a terminal application.

You help users with shell commands, file system tasks, and general programming questions.

You have access to the following tools:
- run_command: execute a shell command in the current working directory
- read_file: read the contents of a file
- list_dir: list the contents of a directory
- get_cwd: get the current working directory

When asked to do something that involves the file system or running commands, prefer using
your tools over explaining how to do it manually.

Keep your responses concise. When you run a command and get the output, summarize what you
found rather than repeating the raw output verbatim.

The user is a developer working in a terminal. Be direct and technical.
```

### 9.8 New Bubble Tea Message Types

One new type was added to `internal/messages/messages.go`:

```go
// AgentConfirmMsg is sent by the agent goroutine when run_command needs
// the user to approve a shell command before it is executed.
// The goroutine blocks on Reply until true (run) or false (cancel) is sent.
type AgentConfirmMsg struct {
    BlockID string
    Command string
    Reply   chan bool
}
```

Tool call notifications (start + name) are sent as `ACPTokenMsg` with an inline `⚙ tool_name…` prefix — no separate `AgentToolCallMsg` or `AgentToolResultMsg` types. The root model renders the confirmation prompt in solar-500 in place of the input bar while `confirmPending != nil`.

### 9.9 Agentic Loop Flow

```
Agent.Send(p, blockID, userInput) returns tea.Cmd
  │
  └── goroutine starts
        │
        ├─ append {role:user, content:userInput} to history
        │
        ├─ LOOP ──────────────────────────────────────────────────────────┐
        │    │                                                             │
        │    ├─ build messages = [systemMsg()] + history                  │
        │    ├─ call client.Chat(ctx, req, onToken)                       │
        │    │       onToken → p.Send(ACPTokenMsg{...})  [0..N times]     │
        │    │                                                             │
        │    ├─ if err → return ACPDoneMsg{Err: err}                      │
        │    │                                                             │
        │    ├─ if resp.ToolCalls == nil:                                  │
        │    │       append {role:assistant, content} to history           │
        │    │       return ACPDoneMsg{BlockID}          ← LOOP EXIT       │
        │    │                                                             │
        │    └─ else (model requested tools):                              │
        │            append {role:assistant, toolCalls} to history        │
        │            for each tool call (sequential):                     │
        │              p.Send(ACPTokenMsg{"\n⚙ tool_name…\n"})           │
        │              if run_command → p.Send(AgentConfirmMsg)          │
        │                              blocks on reply channel           │
        │              result, err = tool.Execute(ctx, args)             │
        │              append {role:tool, result} to history             │
        │            continue LOOP ────────────────────────────────────────┘
```

**Why sequential tool execution?** Parallel tool execution requires synchronising partial history writes and complicates context cancellation. At this scale (one developer, one local model) the latency benefit is negligible — model latency dominates over filesystem ops. Sequential is correct, simple, and debuggable. Parallelism is a named future optimisation, not an oversight.

**Context cancellation:** The goroutine checks `ctx.Done()` between tool calls. If the user presses `ctrl+c`, the cancel propagates to both `client.Chat` (via the HTTP request context) and any in-flight `exec.Run` inside a tool.

**Loop termination:** The loop terminates when either: (a) the model returns a response with no tool calls, (b) `ctx` is cancelled, or (c) `client.Chat` returns an error. There is no explicit turn limit in the MVP. A hard limit (e.g., 10 turns) is a named follow-up to prevent runaway loops.

### 9.10 Changes to Existing Files

| File | Change |
|------|--------|
| `internal/acp/client.go` | Added `Chat(ctx, msgs, tools, onToken)` and `Model() string` methods |
| `internal/acp/types.go` | **New file**: OpenAI wire types (§9.3) |
| `internal/messages/messages.go` | Added `AgentConfirmMsg` |
| `cmd/forge/main.go` | Added `*agent.Agent` field + `confirmFn`; calls `agent.Send(program, blockID, input)`; handles `AgentConfirmMsg` (confirmation prompt in input bar) |
| `internal/agent/agent.go` | **New file**: `Agent`, `Send()`, `Reset()` |
| `internal/agent/tools.go` | **New file**: `Tool` type + 4 MVP tools |
| `internal/agent/prompt.go` | **New file**: `systemPrompt` constant |

`internal/exec`, `internal/session`, `internal/ui/viewport`, `internal/ui/block`, `internal/ui/header`, `internal/ui/input`, `internal/ui/palette` — **no changes**.

### 9.11 What Is Out of Scope for the MVP

| Capability | Reason deferred |
|------------|----------------|
| Parallel tool execution | Complexity vs. negligible latency gain at local-model scale |
| Tool output truncation / summarisation | Let the model handle it; add if context-window errors appear in practice |
| Hard turn limit | Acceptable risk for v1 with a local model; add as a config option later |
| Persistent conversation history | PRD §3 explicitly excludes session history persistence across restarts |
| Tool allowlist / sandbox | Trust-in-local-model decision for v1; document the risk, revisit if remote agents are added |
| ~~Confirmation prompt for destructive commands~~ | **Implemented** via `AgentConfirmMsg` + `confirmPending` in root model |
| Streaming tool call deltas to the UI | Tool calls typically complete in <1s; batching the result is fine for v1 |

---

### ADR-04: `internal/agent` as a New Package (not extending `internal/acp`)

**Status:** Accepted

**Context:**  
Two options exist for where to implement conversation history and the tool dispatch loop: extend `internal/acp/client.go` or create a new `internal/agent` package.

**Decision:**  
New package `internal/agent`.

**Rationale:**

| Concern | Extend `acp` | New `agent` |
|---------|-------------|------------|
| Single responsibility | Violated: transport + orchestration | Maintained: each package does one thing |
| Testability | `Client` becomes a god struct | `Agent` can be tested with a mock `*acp.Client` |
| Import graph | No new edge | `agent → acp` (clean, one direction) |
| Existing test compatibility | `SendPrompt` tests would need to be updated | `SendPrompt` untouched |

**Consequences:**
- (+) `acp.Client` stays focused on HTTP transport; existing tests unchanged.
- (+) `agent.Agent` is independently testable by injecting a fake client.
- (+) Future transports (WebSocket, gRPC) only require changing `acp`; `agent` is transport-agnostic.
- (−) One additional package for contributors to learn. Acceptable at this project size.

---

### ADR-05: OpenAI Wire Types Defined in `internal/acp`, Not a Shared `types` Package

**Status:** Accepted

**Context:**  
`acp.Client.Chat()` and `agent.Agent` both need `Message`, `ToolCall`, `ToolDef`, and `ChatRequest`. Three locations are possible: `internal/acp`, `internal/agent`, or a new `internal/llm` (or `internal/types`) package.

**Decision:**  
Define all wire types in `internal/acp/types.go`.

**Rationale:**  
The types are OpenAI API wire shapes — they belong with the HTTP transport that sends and receives them. `agent` imports `acp` (already required to use `*acp.Client`), so referencing `acp.Message` adds no new dependency edge. A dedicated `types` package would add a third package purely to avoid an import, which is over-engineering for a two-package relationship.

**Consequences:**
- (+) No new package; no new import cycle risk.
- (+) Wire types live next to the code that serialises/deserialises them.
- (−) `acp` package now contains both transport logic and type definitions. This is acceptable because Go packages routinely own both their data types and their behaviour.
