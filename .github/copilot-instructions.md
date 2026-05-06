# Forge — Copilot Workspace Context

**Project**: Forge — a TUI terminal app in Go. A focused opencode clone.

**Stack**: Go 1.23 · Bubble Tea (TUI framework, Elm arch) · Lip Gloss (styling) · ACP over stdio (agent protocol)

## Design Tokens (PRD §10)
| Token | Value | Role |
|-------|-------|------|
| ink-0 | #0a0b10 | Canvas background |
| ink-1 | #0f1118 | Block background |
| ink-2 | #161922 | Header background |
| ember-500 | #ff6b3d | Accent / focused border |
| plasma-500 | #7c5cff | AI card |
| mint-500 | #22c97e | Success |
| crimson-500 | #ef4444 | Failed |
| solar-500 | #eab308 | Running |
| ink-8 | #c5c9d6 | Body text |
| ink-6 | #5a6178 | Muted text |

## Branch Rules
- Feature tasks → branch `T-XX/slug` from main, merge `--no-ff`, delete branch after merge.
- Bug/fix and chore → commit directly to main.

## Task System
- Tasks live in `.github/tasks/`. `index.md` is the single source of truth for status.
- Backlog: `.github/tasks/backlog/` · Active: `.github/tasks/in-progress/` · Done: `.github/tasks/done/`

## Commit Format
Conventional Commits. Scopes: `blocks` · `acp` · `agent` · `input` · `palette` · `config` · `ci` · `tui`

## Before Touching Code
Read the task file in `.github/tasks/in-progress/` and every file listed in its **Tech notes** section.

## Full Docs
- `PRD.md` — requirements and scope
- `DESIGN.md` — design system
- `CONTRIBUTING.md` — workflow
