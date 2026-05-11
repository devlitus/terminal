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

Forge is a TUI terminal application built on Bubble Tea's Elm-style architecture. Every shell command produces a discrete, addressable **Command Block** — a self-contained unit of header, input, and streamed output — rendered in a scrollable viewport between a persistent header bar and input bar. An ACP (Agent Client Protocol) client over stdio connects Forge to a local coding agent, enabling inline AI assistance inside failed blocks without leaving the terminal.

---

## 2. Package Responsibilities

| Package | Import Path | Responsibility |
|---------|-------------|---------------|
| `main` | `github.com/forge-tui/forge/cmd/forge` | Program entry point: initializes config, starts ACP client, constructs the root Bubble Tea model, and calls `tea.NewProgram().Run()` |
| `config` | `github.com/forge-tui/forge/internal/config` | Loads and validates `~/.config/forge/config.toml`; exposes a typed `Config` struct; enables shell-only mode when no agent config is present |
| `theme` | `github.com/forge-tui/forge/internal/theme` | Declares all Lip Gloss `lipgloss.Color` constants mapped from the design token table; the single source of truth for every color value in the TUI |
| `block` | `github.com/forge-tui/forge/internal/block` | Defines the `Block` data model: ID, command string, cwd, state (`Running`/`Success`/`Failed`), exit code, duration, accumulated output bytes, and AI card content |
| `exec` | `github.com/forge-tui/forge/internal/exec` | Starts shell commands via `os/exec` + `io.Pipe`; owns the per-command goroutine that reads stdout/stderr and emits `ExecOutputMsg` and `ExecDoneMsg` as `tea.Cmd` return values |
| `acp` | `github.com/forge-tui/forge/internal/acp` | Manages the ACP subprocess lifecycle over stdio; exposes `Client` which marshals JSON-RPC requests and owns the streaming goroutine that emits `ACPTokenMsg` and `ACPDoneMsg` |
| `session` | `github.com/forge-tui/forge/internal/session` | Tracks current working directory, owns the ordered list of `block.Block` values (the ring buffer), dispatches `CwdChangedMsg` on `cd`, handles `clear` by emitting `ViewportClearMsg` |
| `ui/block` | `github.com/forge-tui/forge/internal/ui/block` | Bubble Tea component for a single Command Block; renders header, command line, and scrollable output; handles focus state and per-block keyboard actions (`r`, `y`, `f`, `ctrl+c`) |
| `ui/header` | `github.com/forge-tui/forge/internal/ui/header` | Renders the 1-line top bar: app name, current working directory, and ACP connection indicator; updates on `CwdChangedMsg` |
| `ui/input` | `github.com/forge-tui/forge/internal/ui/input` | Renders the 1-line bottom input bar; owns the text field; distinguishes shell commands from AI prompts (`/` or `@agent` prefix) and built-in commands (`clear`, `help`); emits `OpenPaletteMsg` on `ctrl+k` |
| `ui/aicard` | `github.com/forge-tui/forge/internal/ui/aicard` | Renders the AI Suggestion Card that appears inside a failed block; streams incoming `ACPTokenMsg` tokens; shows Run/Dismiss actions after `ACPDoneMsg` |
| `ui/viewport` | `github.com/forge-tui/forge/internal/ui/viewport` | Manages the scrollable list of `ui/block` components; routes `ExecOutputMsg`, `ExecDoneMsg`, `ACPTokenMsg`, `ACPDoneMsg`, and `BlockFocusedMsg` to the correct child block by ID; enforces the ring buffer cap (ADR-03) |
| `ui/palette` | `github.com/forge-tui/forge/internal/ui/palette` | Renders the command palette overlay triggered by `ctrl+k`; implements fuzzy filtering over recent commands and built-in actions; emits the selected action as a message on `Enter` |

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

**Sent by:** `internal/acp` — the streaming goroutine, one message per JSON-RPC `stream` event.  
**Consumed by:** `internal/ui/viewport` → `internal/ui/aicard` for the matching `BlockID`.

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

**Sent by:** `internal/acp` — emitted once per `prompt/turn` call.  
**Consumed by:** `internal/ui/aicard` — transitions card from streaming state to Run/Dismiss action state. When `Err != nil`, the card renders a degraded error message (NFR-05).

---

### 3.5 `BlockFocusedMsg`

Signals that keyboard focus has moved to a specific block.

```go
// BlockFocusedMsg is emitted by the viewport when the user navigates
// focus with arrow keys or j/k. An empty BlockID means no block is focused.
type BlockFocusedMsg struct {
    BlockID string
}
```

**Sent by:** `internal/ui/viewport` — on arrow key / `j` / `k` navigation.  
**Consumed by:** `internal/ui/block` components — each component checks whether `BlockID` matches its own ID to set or clear the focused visual state (ember-500 border).

---

### 3.6 `CwdChangedMsg`

Signals that the working directory has changed after a successful `cd` command.

```go
// CwdChangedMsg is emitted by session.Run() when a cd command
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
        │  returns tea.Cmd: session.Run(cmd, cwd)
        ▼
  session.Run()
        │  1. creates block.Block{ID: uuid, State: Running, Cmd: cmd, Cwd: cwd}
        │  2. appends block to ring buffer (enforces N=500 cap)
        │  returns tea.Batch(
        │      BlockCreatedMsg{Block},       ← viewport adds a new ui/block child
        │      exec.Start(blockID, cmd, cwd) ← returns a tea.Cmd
        │  )
        ▼
  exec.Start() → launches goroutine
        │
        │  goroutine: reads io.Pipe (stdout+stderr merged)
        │
        ├── chunk available ──► returns ExecOutputMsg{BlockID, Data}
        │                              │
        │                              ▼
        │                       ui/viewport.Update()
        │                              │ routes by BlockID
        │                              ▼
        │                       ui/block.Update() → appends Data to output buffer
        │                              │
        │                              └─► tea re-renders the block
        │
        └── pipe EOF ──────────► returns ExecDoneMsg{BlockID, ExitCode, Duration}
                                        │
                                        ▼
                                 ui/viewport.Update()
                                        │ routes by BlockID
                                        ▼
                                 ui/block.Update()
                                        │ sets State = Success | Failed
                                        │ sets ExitCode, Duration on header
                                        │
                                        └─► session updates block.Block data model
                                            tea re-renders the block with status badge
```

**Key invariant:** The goroutine in `exec.Start` never writes to shared memory. It communicates exclusively by returning messages through the `tea.Cmd` channel mechanism. The Bubble Tea runtime serialises all `Update` calls on the main goroutine — no locks are needed in any `Update` or `View` function.

---

## 5. Data Flow: ACP / AI Integration

```
User presses 'f' on a focused failed block
        │
        ▼
  ui/block.Update(tea.KeyMsg{"f"})
        │  returns tea.Cmd: acp.FixWithAI(blockID, cmd, output)
        │  also emits BlockFocusedMsg{blockID} to show AI card placeholder
        ▼
  acp.FixWithAI() → launches streaming goroutine
        │  1. marshals prompt: cmd + truncated output + system instructions
        │  2. sends JSON-RPC `prompt/turn` over stdio pipe to agent subprocess
        │
        │  goroutine: reads newline-delimited JSON stream from agent
        │
        ├── token event ──────► returns ACPTokenMsg{BlockID, Token}
        │                              │
        │                              ▼
        │                       ui/viewport.Update()
        │                              │ routes by BlockID
        │                              ▼
        │                       ui/aicard.Update()
        │                              │ appends Token to card content buffer
        │                              └─► tea re-renders card (streaming text)
        │
        └── stream end ────────► returns ACPDoneMsg{BlockID, Err}
                                        │
                                        ▼
                                 ui/aicard.Update()
                                        │ Err == nil → show [Run] [Dismiss] actions
                                        │ Err != nil → show error message, [Dismiss]
                                        └─► tea re-renders card (action state)

  User presses Enter on [Run]
        │
        ▼
  ui/aicard.Update(tea.KeyMsg{"enter"})
        │  extracts suggested command from card content
        │  returns tea.Cmd: session.Run(suggestedCmd, cwd)
        └─► identical flow to §4 shell execution above

  User presses Esc or 'd' on [Dismiss]
        │
        ▼
  ui/aicard.Update() → emits ACPCardDismissedMsg{BlockID}
        └─► ui/viewport removes the AI card from the block, re-renders
```

**Graceful degradation (NFR-05):** If the ACP subprocess is not running or the transport errors, `ACPDoneMsg.Err` is non-nil. The AI card renders: `"Agent unavailable — check config"` with only a Dismiss action. The rest of the TUI continues operating normally.

---

## 6. Concurrency Model

### Goroutines

| Goroutine | Owner | Lifetime | Communicates via |
|-----------|-------|----------|-----------------|
| Bubble Tea event loop | `tea.Program` | Process lifetime | `tea.Msg` channel (internal) |
| Per-command exec reader | `internal/exec` | Duration of one shell command | `tea.Cmd` returning `ExecOutputMsg` / `ExecDoneMsg` |
| ACP streaming reader | `internal/acp` | Duration of one `prompt/turn` call | `tea.Cmd` returning `ACPTokenMsg` / `ACPDoneMsg` |
| ACP subprocess | OS (via `os/exec`) | Process lifetime (managed by `acp.Client`) | stdin/stdout pipes to `acp.Client` |

At peak load there are at most **one exec goroutine per running block** plus **one ACP streaming goroutine**. In practice, Forge runs one command at a time (the input bar is locked while a command runs), so there is typically one exec goroutine active at any moment.

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

### Zone: Header Bar (1 line, top)

**Height:** always 1 terminal row.  
**Content:** `Forge` · `cwd` · ACP connection indicator (dot, coloured mint-500 or crimson-500).

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
**Content:** ember-500 prompt glyph (`❯`) + command text field.

```go
// Lip Gloss style (internal/ui/input)
inputBarStyle := lipgloss.NewStyle().
    Background(lipgloss.Color("#0f1118")). // ink-1
    Width(termWidth).
    Padding(0, 1)

promptStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("#ff6b3d")). // ember-500
    Bold(true)
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

### ADR-02: ACP over stdio transport over HTTP/WebSocket

**Status:** Accepted

**Context:**  
ACP (Agent Client Protocol) supports multiple transports. The agent runs as a local subprocess. Two realistic options are stdio and HTTP (localhost server).

**Decision:**  
Use stdio (stdin/stdout pipes to the agent subprocess) as the ACP transport.

**Rationale:**

| Concern | stdio | HTTP |
|---------|-------|------|
| Network stack required | No | Yes (port binding, firewall, localhost) |
| Subprocess lifetime coupling | Automatic (OS kills agent when TUI exits) | Manual (agent may outlive TUI) |
| Port conflicts | None | Possible (port in use errors) |
| Configuration | Zero (path to agent binary only) | URL + port required |
| Latency | Sub-millisecond IPC | TCP loopback (~0.1ms, negligible but non-zero) |
| Security | Process isolation; no socket exposure | Localhost socket could be probed by other processes |

**Consequences:**
- (+) Zero network configuration; no firewall rules, no port conflicts.
- (+) Agent subprocess is automatically cleaned up when the TUI exits (OS pipe close cascades to the subprocess's stdin read returning EOF).
- (+) No socket exposure reduces attack surface.
- (−) Stdio transport means only one agent per TUI process. Multi-agent scenarios (out of scope for v1 per PRD §12) would require a different transport.
- (−) Debugging the ACP wire protocol requires intercepting the pipe (e.g., `tee`), unlike HTTP which can be inspected with curl.

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

> **Status:** Design · May 2026  
> This section specifies the architecture for evolving the current single-turn ACP client into a stateful, tool-calling agent. It is the canonical reference for the implementation task.

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
    client.go    — EXTEND: add Chat() method alongside existing SendPrompt
    types.go     — NEW: OpenAI-compatible wire types (Message, ToolCall, ToolDef, …)
  agent/
    agent.go     — NEW: Agent struct, Send(), Reset()
    tools.go     — NEW: Tool type + 4 MVP tool implementations
    prompt.go    — NEW: system prompt builder
  messages/
    messages.go  — EXTEND: add AgentToolCallMsg and AgentToolResultMsg
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
// Chat sends req to the OpenAI-compatible /chat/completions endpoint.
// onToken is called for each streamed content delta (may not be called at all
// if the model produces only tool calls). Chat blocks until the stream ends.
// The returned ChatResponse is assembled from the full stream.
func (c *Client) Chat(ctx context.Context, req ChatRequest, onToken func(string)) (*ChatResponse, error)

// Model returns the configured model name (needed by agent to build ChatRequest).
func (c *Client) Model() string
```

`Chat` shares the SSE parsing logic with `sendPromptHTTP`; the delta struct is extended to also capture `tool_calls` deltas per the OpenAI streaming format. Tool call deltas are accumulated and assembled into `ChatResponse.ToolCalls` after the stream ends.

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

// New constructs an Agent with the four MVP tools pre-registered.
func New(client *acp.Client, cwd func() string) *Agent

// Send starts the agentic loop for userInput as a tea.Cmd.
// The returned cmd runs in a goroutine. It emits intermediate messages via
// p.Send() and returns ACPDoneMsg as its final tea.Msg when the loop ends.
func (a *Agent) Send(p *tea.Program, blockID, userInput string) tea.Cmd

// Reset clears conversation history. Called when the user runs `clear`.
func (a *Agent) Reset()
```

`history` is a flat `[]acp.Message`. A system message is prepended on every `Chat` call (not stored in history) so that the system prompt always reflects the current CWD.

### 9.7 System Prompt (`internal/agent/prompt.go`)

```go
func buildSystemPrompt(cwd, shell string) string
```

Content:

```
You are Forge, an AI assistant embedded in a terminal application.
You help developers run commands, navigate the filesystem, read code, and debug problems.

Current working directory: {cwd}
Shell: {shell}

Guidelines:
- When the user asks you to do something, call the appropriate tool directly. Do not describe what you would do — do it.
- Keep prose responses short. Developers read output, not essays.
- Before running a destructive command (rm, git reset --hard, git push --force, anything with sudo), explain what it does and ask for confirmation.
- When a command produces long output, summarise it; do not echo thousands of lines verbatim.
- Prefer non-interactive command variants (--no-pager, --no-edit, -y flags where safe).
```

CWD is injected at every request (not stored in history) so it reflects `cd` changes without needing a special update path.

### 9.8 New Bubble Tea Message Types

Two new types are added to `internal/messages/messages.go`:

```go
// AgentToolCallMsg is emitted when the agent begins executing a tool call
// requested by the model. Used to update the AI card with a status line.
type AgentToolCallMsg struct {
    BlockID  string
    ToolName string
    Args     string // raw JSON arguments, for display only
}

// AgentToolResultMsg is emitted when a tool call completes.
// Err is non-nil if the tool itself failed (distinct from the model failing).
type AgentToolResultMsg struct {
    BlockID  string
    ToolName string
    Result   string
    Err      error
}
```

The root model and viewport forward these to the AI card. The AI card renders a compact status line (e.g., `⚙ run_command: go build ./...`) while the tool is running, replaced by the result summary when done. This is additive — no existing message handling changes.

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
        │              p.Send(AgentToolCallMsg{...})                      │
        │              result, err = tool.Execute(ctx, args)              │
        │              p.Send(AgentToolResultMsg{...})                    │
        │              append {role:tool, result} to history              │
        │            continue LOOP ────────────────────────────────────────┘
```

**Why sequential tool execution?** Parallel tool execution requires synchronising partial history writes and complicates context cancellation. At this scale (one developer, one local model) the latency benefit is negligible — model latency dominates over filesystem ops. Sequential is correct, simple, and debuggable. Parallelism is a named future optimisation, not an oversight.

**Context cancellation:** The goroutine checks `ctx.Done()` between tool calls. If the user presses `ctrl+c`, the cancel propagates to both `client.Chat` (via the HTTP request context) and any in-flight `exec.Run` inside a tool.

**Loop termination:** The loop terminates when either: (a) the model returns a response with no tool calls, (b) `ctx` is cancelled, or (c) `client.Chat` returns an error. There is no explicit turn limit in the MVP. A hard limit (e.g., 10 turns) is a named follow-up to prevent runaway loops.

### 9.10 Changes to Existing Files

| File | Change |
|------|--------|
| `internal/acp/client.go` | Add `Chat(ctx, req, onToken)` and `Model() string` methods |
| `internal/acp/types.go` | **New file**: OpenAI wire types (§9.3) |
| `internal/messages/messages.go` | Add `AgentToolCallMsg` and `AgentToolResultMsg` |
| `cmd/forge/main.go` | Replace `*acp.Client` field with `*agent.Agent`; replace `startACPStream(...)` calls with `agent.Send(program, blockID, input)` |
| `internal/ui/aicard/model.go` | Handle `AgentToolCallMsg` and `AgentToolResultMsg` to render tool status lines |

`internal/exec`, `internal/session`, `internal/ui/viewport`, `internal/ui/block`, `internal/ui/header`, `internal/ui/input`, `internal/ui/palette` — **no changes**.

### 9.11 What Is Out of Scope for the MVP

| Capability | Reason deferred |
|------------|----------------|
| Parallel tool execution | Complexity vs. negligible latency gain at local-model scale |
| Tool output truncation / summarisation | Let the model handle it; add if context-window errors appear in practice |
| Hard turn limit | Acceptable risk for v1 with a local model; add as a config option later |
| Persistent conversation history | PRD §3 explicitly excludes session history persistence across restarts |
| Tool allowlist / sandbox | Trust-in-local-model decision for v1; document the risk, revisit if remote agents are added |
| Confirmation prompt for destructive commands | High-value safety feature but requires new UI; schedule as a follow-up task |
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
