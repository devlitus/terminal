# Forge — Task Index

> Single source of truth for task status. Updated by the Orchestrator only.

## Legend
| Status | Meaning |
|--------|---------|
| 🔲 backlog | Ready to start — dependencies met |
| ▶️ in-progress | Actively being worked on |
| ✅ done | Merged to main |
| ⏸ blocked | Waiting on a dependency |

## Tasks

| ID | Title | Agent | Status | Depends on |
|----|-------|-------|--------|------------|
| T-01 | Project Scaffold | Code Expert | ✅ done | — |
| T-02 | Architecture Document | Architect | ✅ done | T-01 |
| T-03 | Design Tokens Package | Code Expert | ✅ done | T-01, T-02 |
| T-04 | Block Data Model | Code Expert | ✅ done | T-01, T-02 |
| T-05 | Config Package | Code Expert | ✅ done | T-01, T-02 |
| T-06 | Shell Executor | Code Expert | ✅ done | T-01, T-02 |
| T-07 | ACP Client | Code Expert | ✅ done | T-02, T-05 |
| T-08 | Command Block UI Component | Code Expert | ✅ done | T-03, T-04 |
| T-09 | Header Bar Component | Code Expert | ✅ done | T-03 |
| T-10 | Input Bar Component | Code Expert | ✅ done | T-03 |
| T-11 | AI Suggestion Card Component | Code Expert | ✅ done | T-03, T-04 |
| T-12 | Block Viewport | Code Expert | ✅ done | T-08 |
| T-13 | Shell Execution Integration | Code Expert | ✅ done | T-06, T-10, T-12 |
| T-14 | Block Actions — Copy and Re-run | Code Expert | ✅ done | T-12, T-13 |
| T-15 | Session Management | Code Expert | ✅ done | T-09, T-13 |
| T-16 | ACP / AI Integration | Code Expert | ✅ done | T-07, T-11, T-13 |
| T-17 | Command Palette | Code Expert | ✅ done | T-10, T-12 |
| T-18 | Root TUI Assembly | Code Expert | ✅ done | T-14, T-15, T-16, T-17 |
| T-19 | Graceful ACP Degradation | Code Expert | ✅ done | T-16, T-18 |
| T-20 | Performance and Regression Tests | Code Expert | 🔲 backlog | T-18 |
| T-21 | CI Finalization and README | Code Expert | 🔲 backlog | T-18 |
