# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

**Forge** is a TUI terminal app in Go — a focused [opencode](https://opencode.ai) clone. Every shell command is rendered as an addressable **Command Block** (header + input + output). Failed blocks can trigger an agentic AI loop that streams a fix suggestion inline.

Forge degrades gracefully: if `~/.config/forge/config.toml` is missing or the ACP endpoint is unreachable, it runs as a plain shell TUI with no AI features.

## Commands

```bash
# Build
go build -o forge ./cmd/forge

# Run
go run ./cmd/forge

# Test (always use race detector)
go test -race ./...
go test -race ./internal/block/...   # single package

# Lint
golangci-lint run
```

## Architecture

The entry point is `cmd/forge/main.go`, which owns `rootModel` — the root Bubble Tea model. All state mutations are serialised through Bubble Tea's `Update()` loop; no locks are needed.

**Package responsibilities:**

| Package | Role |
|---|---|
| `internal/messages` | All Bubble Tea message types. No deps — exists solely to break import cycles. |
| `internal/block` | `Block` data model + `RingBuffer` (cap 500, front-drop eviction). |
| `internal/session` | Tracks CWD; interprets built-ins (`cd`, `clear`). |
| `internal/exec` | Spawns per-command goroutines; pipes stdout/stderr as `ExecOutputMsg` / `ExecDoneMsg`. |
| `internal/acp` | HTTP transport to any OpenAI-compatible endpoint; SSE streaming. |
| `internal/agent` | Agentic loop on top of `acp`: conversation history, tool dispatch (`run_command`, `read_file`, `list_dir`, `get_cwd`). |
| `internal/config` | Loads `~/.config/forge/config.toml` via BurntSushi/toml. |
| `internal/theme` | Lip Gloss color tokens — single source of truth for all colors. |
| `internal/ui/*` | One sub-package per visual component: `block`, `viewport`, `aicard`, `header`, `input`, `palette`. |

**Concurrency rule:** Background goroutines (`exec`, `acp`) never write to shared state. They communicate exclusively via Bubble Tea messages sent to the event loop.

**Input modes:** The bottom input bar has two modes toggled by `!`:
- Shell mode — input is executed as a command.
- Chat mode — input is sent to the agent (`/` and `@` also enter chat mode).

## Design tokens

Colors are defined in `internal/theme`. Use the token names, not raw hex values:

| Token | Role |
|---|---|
| `ember-500` `#ff6b3d` | Accent / focused border |
| `plasma-500` `#7c5cff` | AI card |
| `mint-500` `#22c97e` | Success |
| `crimson-500` `#ef4444` | Failed |
| `solar-500` `#eab308` | Running |

## Workflow

- Feature work → branch `T-XX/slug` from main, merge `--no-ff`, delete after merge.
- Bug/fix/chore → commit directly to main.
- Commit format: Conventional Commits. Scopes: `blocks` · `acp` · `agent` · `input` · `palette` · `config` · `ci` · `tui`.
- Before touching code, read the active task file in `.github/tasks/in-progress/` and every file listed in its **Tech notes** section.

## Agent team

Specialists live in `.claude/agents/` and are invoked automatically based on the nature of the request. You decide when to delegate — do not wait for the user to ask.

| Specialist | When to delegate |
|---|---|
| `architect` | Structural decisions, package boundaries, new patterns, trade-off analysis, ADRs |
| `code-expert` | Writing, reviewing, or refactoring Go code; implementing task acceptance criteria |
| `security` | Any code touching subprocess spawning, config/key handling, user input → shell |
| `testing` | Writing or reviewing `*_test.go` files in `internal/` |
| `code-review` | Final review after implementation: checks all acceptance criteria, Forge conventions, and code quality; flags blockers and suggests improvements |

**Orchestration workflows:**

_New feature_ — all agents participate in order:
0. Create branch `T-XX/slug` from main before any code change.
1. `architect` — design, package boundaries, ADR if needed.
2. `code-expert` — implementation against the architect's spec.
3. `code-review` — verify all acceptance criteria and conventions before continuing.
4. `security` — audit any security surface in the new code.
5. `testing` — write or update `*_test.go` files.

_Bug / fix / chore_ — minimal team:
1. `code-expert` — diagnose and fix.
2. `code-review` — verify the fix is correct and complete.
3. `security` — only if the fix touches subprocess spawning, config/key handling, or user input → shell.
4. `testing` — only if an existing test must be modified or a regression test is needed.

**General rules:**
- Vague request → ask one clarifying question before delegating.
- Synthesize specialist output — never relay it raw. Surface conflicts and state your recommendation.

**Utilities (manual slash commands):**
- `/new-adr <kebab-title>` — create an Architecture Decision Record

**Task management** is handled automatically by the `architect` agent via the `task` skill. The architect creates tasks when a design decision produces actionable work; it also moves tasks between backlog → in-progress → done.

## Key reference docs

- `ARCHITECTURE.md` — canonical reference: message types, data flows, concurrency model, ADRs.
- `DESIGN.md` — full design system.
- `CONTRIBUTING.md` — workflow details and CI checks.
- `PRD.md` — product requirements and scope.
