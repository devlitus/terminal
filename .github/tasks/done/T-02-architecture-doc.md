---
id: T-02
status: backlog
agent: Architect
depends-on: [T-01]
branch: T-02/architecture-doc
---

## T-02 · Architecture Document

**Context**
Defines the package layout, data flow, key interfaces, and component boundaries before any significant code is written. All implementation tasks cite this as the authoritative reference for types, message names, and package responsibilities.

**Acceptance criteria**
- [ ] `ARCHITECTURE.md` committed to the repo root
- [ ] Documents the top-level package tree with a one-line responsibility statement per package
- [ ] Defines all cross-boundary Bubble Tea message types (minimum: `ExecOutputMsg`, `ExecDoneMsg`, `ACPTokenMsg`, `ACPDoneMsg`, `BlockFocusedMsg`, `CwdChangedMsg`, `OpenPaletteMsg`, `ViewportClearMsg`)
- [ ] Describes the data flow: user keystroke → shell exec → block update → re-render
- [ ] Describes the ACP integration flow: user trigger → ACP request → streaming response → AI card update
- [ ] Identifies the concurrency boundary: which goroutines exist, which channels they use, how results are fed back into Bubble Tea via `tea.Cmd`
- [ ] Names the three Lip Gloss layout zones: header bar, block viewport, input bar — with height allocations
- [ ] Includes a trade-offs section: why Elm-arch Bubble Tea over direct termbox, why stdio ACP transport over HTTP
- [ ] Includes a short ADR for the ring buffer cap decision (PRD OQ-2)

**Tech notes**
- Read PRD §5 (core concepts), §8 (stack), §9 (ACP methods used), §12 (explicit out of scope) before drafting
- Read `DESIGN.md` §6 (Components) for UI component anatomy
- Output must be `ARCHITECTURE.md` at the repo root — not in chat, not in a subdirectory
- No implementation code — describe contracts, types, and responsibilities only
- Reference PRD sections inline (e.g., "FR-02 requires streaming stdout: …")
- This document is the single source of truth for message type names; all subsequent tasks must use the names defined here

**Completion notes**
*(empty)*
