# Forge — Product Requirements Document

> Version 1.0 · May 2026 · Status: v1 Implemented

---

## 1. Overview

**Forge** is a terminal UI (TUI) for AI-assisted development. It presents every command execution as a discrete, addressable block — header, input, output — and connects to any OpenAI-compatible HTTP endpoint to provide inline AI assistance.

Forge is a focused clone of opencode. It ships only the primitives that matter, without the enterprise surface area.

---

## 2. Goals

| # | Goal |
|---|------|
| G1 | A terminal that feels native but renders command output as structured blocks |
| G2 | Inline AI assistance that lives inside the terminal, not beside it |
| G3 | Every block is addressable: shareable, replayable, focusable |
| G4 | Zero configuration to start; connects to any OpenAI-compatible HTTP endpoint |

---

## 3. Non-goals (v1)

- Multi-tab / multi-pane layout
- Remote agent execution (cloud-hosted agents)
- Auth / team sharing / URL-based block sharing
- Custom themes or Tweaks panel
- Voice input
- Diff views or inline file editing
- Session history persistence across restarts
- Windows native support (WSL is the supported path on Windows)

---

## 4. Users

**Primary:** Solo developer running CLI-heavy workflows (git, npm, docker, deploy scripts) who wants AI assistance without leaving the terminal.

**Not targeting:** Teams, DevOps pipelines, non-technical users.

---

## 5. Core concepts

### 5.1 Command Block

The atomic unit of the UI. Every executed command produces one block.

```
┌─ Header ──────────────────────────────────────────────────────┐
│  ~/forge/api  ·  exit 0  ·  1.2s              [success]  [⋯] │
├─ Input ────────────────────────────────────────────────────────┤
│  ❯  git status                                                 │
├─ Output ───────────────────────────────────────────────────────┤
│  On branch feat/blocks                                         │
│  Your branch is up to date with 'origin/feat/blocks'.         │
│  ...                                                           │
└────────────────────────────────────────────────────────────────┘
```

**Block states:**

| State    | Badge         | Actions available         |
|----------|---------------|---------------------------|
| running  | `in progress` | Stop                      |
| success  | `success`     | Re-run, Copy, Focus       |
| failed   | `failed`      | Fix with AI, Re-run, Copy |

### 5.2 AI Suggestion Card

Appears inside a block after "Fix with AI" is triggered. Contains explanation, proposed command, and Run/Dismiss actions.

### 5.3 Input Bar

Always visible at the bottom. Has two modes toggled by `!`:

- **Chat mode** (default, `⬡` prompt) — input is sent to the AI agent. `/` and `@` also switch to chat mode from shell mode.
- **Shell mode** (`❯` prompt) — input is executed as a shell command. `!command` executes a one-off shell command without permanently switching mode.

Built-in commands (`clear`, `ctrl+k`) are intercepted regardless of mode.

---

## 6. Functional requirements

### 6.1 Shell execution
- FR-01: Execute arbitrary shell commands in the current working directory
- FR-02: Stream stdout/stderr output into the block in real time; ANSI/VT100 escape sequences that manipulate cursor or screen are stripped before rendering
- FR-03: Capture exit code and wall-clock duration per block
- FR-04: Support `ctrl+c` to kill a running command

### 6.2 Command blocks
- FR-05: Each command produces exactly one block
- FR-06: Blocks scroll vertically; the viewport shows the most recent N blocks
- FR-07: A focused block is visually distinguished (accent border)
- FR-08: Blocks can be focused with keyboard (`up / down` arrow keys)
- FR-09: Copy the output of a focused block (`ctrl+y`)
- FR-10: Re-run the command in a focused block (`ctrl+r`)

### 6.3 AI assistance
- FR-11: Trigger "Fix with AI" on any failed block (`ctrl+f` or via palette)
- FR-12: Send a free-form prompt to the agent from the input bar (chat mode)
- FR-13: Agent response streams token-by-token into an AI card; tool calls are shown inline as `⚙ tool_name…`
- FR-14: Accept an AI-suggested command with one keystroke (`Enter` on the card)
- FR-15: Dismiss an AI card without running (`Esc` or `ctrl+d`)

### 6.4 Session
- FR-16: Session starts at the current working directory
- FR-17: `cd` is a session built-in (no subprocess); changes the working directory and the next block header reflects it
- FR-18: `clear` is a session built-in (no subprocess); resets the ring buffer and clears the viewport

### 6.5 Command palette
- FR-19: `ctrl+k` opens a fuzzy-search palette over recent commands and built-in actions
- FR-20: Palette items: re-run command, copy output, fix with AI

---

## 7. Non-functional requirements

| # | Requirement |
|---|-------------|
| NFR-01 | First paint < 100ms on a modern laptop |
| NFR-02 | Keystroke-to-character latency < 16ms |
| NFR-03 | No memory leaks on sessions > 1 hour |
| NFR-04 | Works at terminal widths 80–220 columns |
| NFR-05 | AI endpoint unavailable or unreachable must not crash the TUI; degrade gracefully |

---

## 8. Technical stack

| Layer | Choice | Reason |
|-------|--------|--------|
| Language | Go 1.23+ | Performance, goroutines for concurrent blocks, stdlib covers all protocol needs |
| TUI framework | Bubble Tea | Elm architecture, native viewport/scroll, first-class in Go TUI ecosystem |
| Styling | Lip Gloss | CSS-like layout, borders, colors — maps directly to the Forge design tokens |
| Agent protocol | OpenAI-compatible HTTP + SSE | Direct HTTP to any OpenAI-compatible endpoint; streaming via Server-Sent Events |
| Shell execution | os/exec + io.Pipe | No dependency needed; goroutines handle concurrent streaming |

---

## 9. Protocol integration (HTTP)

Forge calls any **OpenAI-compatible HTTP endpoint** directly — no subprocess, no ACP. The `internal/acp` package wraps `POST /chat/completions` with SSE streaming.

### Connection

If `api_base` is empty in config, Forge starts in **shell-only mode** — no AI features, no crash. The header bar shows an inline hint in ink-6:
> `AI offline — add config: ~/.config/forge/config.toml`

### Agentic loop (`internal/agent`)

The `agent` package maintains a multi-turn conversation history and dispatches tool calls requested by the model before sending the results back for a follow-up reply. Tools available to the agent:

| Tool | What it does |
|---|---|
| `run_command` | Executes a shell command (user confirmation required) |
| `read_file` | Reads a file from the filesystem |
| `list_dir` | Lists directory contents |
| `get_cwd` | Returns the current working directory |

### Graceful degradation (NFR-05)

- If `api_base` is empty or `Connect` fails, `shellOnly = true` is set at startup.
- The root model sets `shellOnly = true` on any unrecoverable HTTP error mid-session.
- AI cards show `"AI error — try again"` in crimson-500 when a prompt fails.
- `ctrl+f` in shell-only mode shows the offline hint in the header instead of calling the API.

---

## 10. Design system

Forge uses the design tokens and component patterns defined in `DESIGN.md`.

### Token mapping: Web to TUI

| Design token | Value | Lip Gloss use |
|---|---|---|
| `--ink-0` | `#0a0b10` | canvas background |
| `--ink-1` | `#0f1118` | block background |
| `--ink-2` | `#161922` | header background |
| `--ember-500` | `#ff6b3d` | primary accent, focused block border |
| `--plasma-500` | `#7c5cff` | AI card surfaces |
| `--mint-500` | `#22c97e` | success badge |
| `--crimson-500` | `#ef4444` | failed badge |
| `--solar-500` | `#eab308` | running / warning |
| `--ink-8` | `#c5c9d6` | body text |
| `--ink-6` | `#5a6178` | muted text (metadata) |
| `--syn-keyword` | `#ff8a6a` | shell command keyword highlight |
| `--syn-string` | `#4ade94` | quoted strings in output |
| `--syn-flag` | `#f87171` | CLI flags |

### Block anatomy in TUI

```
Header:  bg=ink-2   fg=ink-7   badge=status-dependent   height=1 line
Input:   bg=ink-1   prompt=ember-500   command=ink-9    height=1 line
Output:  bg=ink-0   fg=ink-8   scrollable               max-height=terminal/3
Border:  default=ink-3   focused=ember-500   (rounded, 1 char each side)
```

Inner content renders at `terminalWidth - 2` to account for the rounded border characters. This ensures blocks stay within the terminal at any width (NFR-04).

ANSI/VT100 escape sequences (cursor movement, screen clears) are stripped from command output before rendering to prevent TUI corruption.

### Root layout

```
┌─ Header bar ────────────────────────────────┐  height = 1 line (+ 1 if statusHint)
├─ Block viewport ────────────────────────────┤  height = termHeight - 2
└─ Input bar ─────────────────────────────────┘  height = 1 line
```

Launched with `tea.WithAltScreen()` for full terminal ownership.

---

## 11. Keyboard shortcuts

| Key | Action |
|-----|--------|
| `Enter` | Execute command / accept AI card |
| `up / down` | Focus previous / next block |
| `!` | Toggle between chat mode and shell mode |
| `!command` | Execute a one-off shell command from chat mode |
| `ctrl+f` | Fix with AI (on focused failed block) |
| `ctrl+r` | Re-run focused block |
| `ctrl+y` | Copy focused block output to clipboard |
| `ctrl+c` | Kill running command; quits if no command is running |
| `ctrl+q` | Quit immediately if no commands running; prompts `Quit? (y/n)` if any block is `StateRunning` |
| `ctrl+k` | Open command palette |
| `Esc` | Dismiss AI card / close palette / cancel quit prompt |
| `ctrl+d` | Dismiss AI card (alternative to `Esc`) |
| `y / n` | Confirm or cancel quit prompt / agent tool confirmation |

---

## 12. Explicit out of scope for v1

These will NOT be built and should not influence architecture decisions:

- File tree / explorer panel
- Git diff viewer
- Split panes
- Remote / cloud agent connections
- Plugin system
- Session persistence (SQLite, etc.)
- Themes switcher
- Windows native binary (WSL only)
- Web UI companion

---

## 13. Success criteria for v1

- [x] A developer can open Forge, run 10 commands, and see them as scrollable blocks
- [x] A failed command shows "Fix with AI"; the suggestion runs with one keystroke
- [x] The TUI connects to any OpenAI-compatible HTTP endpoint without additional tooling
- [x] No crash on 500+ blocks in a session (ring buffer capped at 500, performance tests pass)
- [x] Keyboard-only workflow is fully functional (no mouse required)

---

## 14. Decisions (resolved)

| # | Question | Decision |
|---|----------|----------|
| OQ-1 | Which agent backend ships by default? | Direct HTTP to any OpenAI-compatible endpoint (`api_base` in config). No subprocess needed. |
| OQ-2 | Cap blocks in memory? | Ring buffer, cap at N=500. Drop oldest when exceeded. |
| OQ-3 | Minimum Go version? | Go 1.23. |
| OQ-4 | How is shell-only mode triggered? | `api_base` empty in config, or any unrecoverable HTTP error mid-session. |
| OQ-5 | How are `cd` and `clear` handled? | Intercepted by the `session` package before reaching the shell; no subprocess spawned. |

### OQ-1 detail — Agent backend

Forge calls the OpenAI-compatible `/chat/completions` endpoint directly via HTTP + SSE. Credentials are passed via config:

```toml
# ~/.config/forge/config.toml
[agent]
api_base = "http://localhost:11434/v1"  # Ollama (default, free)
api_key  = ""                           # empty for local models
model    = "qwen2.5-coder:7b"
```

Defaults point to a local Ollama instance.

**Recommended models:**

| Use case | Model | Cost |
|---|---|---|
| Local, offline, free | Ollama + `qwen2.5-coder:7b` | Free |
| Best quality, cheap | DeepSeek V3 | ~$0.07/M tokens |
| Complex reasoning | DeepSeek R1 | ~$0.55/M tokens |
| Maximum flexibility | OpenRouter | Variable |

If the agent binary is not found or `api_base` is empty, Forge starts in **shell-only mode** and shows an inline setup hint in the header.
