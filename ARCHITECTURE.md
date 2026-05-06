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
