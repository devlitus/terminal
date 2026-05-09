# Forge — Product Requirements Document

> Version 1.0 · May 2026 · Status: v1 Implemented

---

## 1. Overview

**Forge** is a terminal UI (TUI) for AI-assisted development. It presents every command execution as a discrete, addressable block — header, input, output — and connects to a coding agent via the Agent Client Protocol (ACP) to provide inline AI assistance.

Forge is a focused clone of opencode. It ships only the primitives that matter, without the enterprise surface area.

---

## 2. Goals

| # | Goal |
|---|------|
| G1 | A terminal that feels native but renders command output as structured blocks |
| G2 | Inline AI assistance that lives inside the terminal, not beside it |
| G3 | Every block is addressable: shareable, replayable, focusable |
| G4 | Zero configuration to start; connects to any ACP-compatible agent |

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

Always visible at the bottom. Handles shell command execution, AI prompts (prefix with `/` or `@agent`), and built-in commands (`clear`, `help`, `ctrl+k`).

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
- FR-08: Blocks can be focused with keyboard (arrow keys / `j/k`)
- FR-09: Copy the command or output of a focused block (`y`)
- FR-10: Re-run the command in a focused block (`r`)

### 6.3 AI assistance (ACP)
- FR-11: Trigger "Fix with AI" on any failed block (`f` key or button)
- FR-12: Send a free-form prompt to the agent from the input bar
- FR-13: Agent response streams token-by-token into an AI card
- FR-14: Accept an AI-suggested command with one keystroke (`Enter` on the card)
- FR-15: Dismiss an AI card without running (`Esc` or `d`)

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
| NFR-05 | ACP agent connection failure must not crash the TUI; degrade gracefully |

---

## 8. Technical stack

| Layer | Choice | Reason |
|-------|--------|--------|
| Language | Go 1.23+ | Performance, goroutines for concurrent blocks, stdlib covers all protocol needs |
| TUI framework | Bubble Tea | Elm architecture, native viewport/scroll, first-class in Go TUI ecosystem |
| Styling | Lip Gloss | CSS-like layout, borders, colors — maps directly to the Forge design tokens |
| Agent protocol | ACP (stdio transport) | Designed for editor↔agent; has a Terminals primitive that matches command blocks exactly |
| Shell execution | os/exec + io.Pipe | No dependency needed; goroutines handle concurrent streaming |

---

## 9. Protocol integration (ACP)

Forge acts as an **ACP Client**. The coding agent runs as a local subprocess (ACP Server over stdio).

### Agent subprocess launch

The agent binary is resolved in order:
1. `FORGE_AGENT_CMD` environment variable (if set)
2. `forge-agent` on `$PATH`

If the binary is not found, or if `api_base` is empty in config, Forge starts in **shell-only mode** — no AI features, no crash. The header bar shows an inline hint in ink-6:
> `AI offline — add config: ~/.config/forge/config.toml`

### Graceful degradation (NFR-05)

- A background goroutine calls `cmd.Wait()` to detect subprocess death. After death, `SendPrompt` returns `ErrNotConnected` without blocking.
- The root model sets `shellOnly = true` on any unrecoverable ACP error mid-session.
- AI cards show `"AI error — try again"` in crimson-500 when a prompt fails.
- `f` key in shell-only mode shows the offline hint in the header instead of calling ACP.

### ACP messages used in v1

| Message | When Forge uses it |
|--------|-----------------|
| `session/create` | On startup |
| `prompt/turn` | User sends a message or triggers "Fix with AI" |
| streaming tokens | Agent response streams token-by-token into the AI card |
| `session/close` | On quit |

### Out of scope for v1
`session/list`, `session/resume`, `session/fork`, `slash_commands`, auth methods, `tool_calls` display, `terminals/*`.

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
| `Enter` | Execute command |
| `up / down` or `j / k` | Focus previous / next block |
| `f` | Fix with AI (on failed block) |
| `r` | Re-run focused block |
| `y` | Copy focused block output |
| `ctrl+c` | Kill running command |
| `ctrl+k` | Open command palette |
| `Esc` | Dismiss AI card / close palette |
| `q` | Quit immediately if no commands running; prompts `Quit? (y/n)` if any block is `StateRunning` |
| `ctrl+c` (root) | Cancels the running command (does not quit); quits if no command is running |

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
- [x] The TUI connects to an ACP-compatible agent on startup without configuration
- [x] No crash on 500+ blocks in a session (ring buffer capped at 500, performance tests pass)
- [x] Keyboard-only workflow is fully functional (no mouse required)

---

## 14. Decisions (resolved)

| # | Question | Decision |
|---|----------|----------|
| OQ-1 | Which agent backend ships by default? | External `forge-agent` subprocess over ACP/stdio. Config passes API credentials. |
| OQ-2 | Cap blocks in memory? | Ring buffer, cap at N=500. Drop oldest when exceeded. |
| OQ-3 | Minimum Go version? | Go 1.23. |
| OQ-4 | How is the agent binary resolved? | `FORGE_AGENT_CMD` env var → `forge-agent` on `$PATH` → shell-only fallback. |
| OQ-5 | How are `cd` and `clear` handled? | Intercepted by the `session` package before reaching the shell; no subprocess spawned. |

### OQ-1 detail — Agent backend

Forge connects to an external `forge-agent` subprocess that acts as both ACP server and LLM client. The agent speaks the OpenAI-compatible API; credentials are passed via config:

```toml
# ~/.config/forge/config.toml
[agent]
api_base = "http://localhost:11434/v1"  # Ollama (default, free)
api_key  = ""                           # empty for local models
model    = "qwen2.5-coder:7b"
```

Defaults point to a local Ollama instance. Forge itself never makes HTTP calls — it only speaks ACP over stdio to the agent subprocess.

**Recommended models:**

| Use case | Model | Cost |
|---|---|---|
| Local, offline, free | Ollama + `qwen2.5-coder:7b` | Free |
| Best quality, cheap | DeepSeek V3 | ~$0.07/M tokens |
| Complex reasoning | DeepSeek R1 | ~$0.55/M tokens |
| Maximum flexibility | OpenRouter | Variable |

If the agent binary is not found or `api_base` is empty, Forge starts in **shell-only mode** and shows an inline setup hint in the header.
