# Forge

Forge is a terminal UI (TUI) for AI-assisted development. It renders every command execution as a discrete, addressable block — header, input, output — and connects to a coding agent via the Agent Client Protocol (ACP) to provide inline AI assistance directly inside the terminal. Forge is a focused clone of [opencode](https://opencode.ai), shipping only the primitives that matter.

See [PRD.md](PRD.md) for full requirements and scope.

---

## Prerequisites

| Requirement | Version |
|-------------|---------|
| Go | 1.23 or later |
| Linux / macOS | — |
| Windows | WSL 2 (native Windows binary is not supported in v1) |

---

## Build

```sh
git clone https://github.com/forge-tui/forge.git
cd forge
go build ./cmd/forge -o forge
```

The resulting `forge` binary has no runtime dependencies.

---

## Run

```sh
./forge
```

Forge looks for an agent config at `~/.config/forge/config.toml`. If none is found it starts in **shell-only mode** (see below).

### Optional: connect an AI agent

Create `~/.config/forge/config.toml`:

```toml
[agent]
api_base = "http://localhost:11434/v1"  # any OpenAI-compatible endpoint
api_key  = ""                           # leave empty for local models
model    = "qwen2.5-coder:7b"           # any model served by your endpoint
```

Recommended models:

| Use case | Model |
|----------|-------|
| Local, offline, free | Ollama + `qwen2.5-coder:7b` |
| Best quality, cheap  | DeepSeek V3 |
| Complex reasoning    | DeepSeek R1 |

---

## Shell-only mode

If no config file exists, or the agent endpoint is unreachable at startup, Forge runs as a plain TUI shell. All command execution, block navigation, copy, and re-run features work normally. AI features (Fix with AI, free-form prompts, AI cards) are disabled and a one-line hint is shown at the bottom of the screen. No crash, no prompt — Forge degrades gracefully.

---

## Keyboard shortcuts

| Key | Action |
|-----|--------|
| `Enter` | Execute command |
| `↑ / ↓` or `j / k` | Focus previous / next block |
| `f` | Fix with AI (failed block) |
| `r` | Re-run focused block |
| `y` | Copy focused block output |
| `ctrl+c` | Kill running command |
| `ctrl+k` | Open command palette |
| `Esc` | Dismiss AI card / close palette |
| `q` | Quit |

---

## Tests

```sh
go test -race ./...
```
