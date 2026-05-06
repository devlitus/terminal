---
id: T-03
status: backlog
agent: Code Expert
depends-on: [T-01, T-02]
branch: T-03/theme-tokens
---

## T-03 · Design Tokens Package

**Context**
Centralises all Lip Gloss color constants and pre-built style values so every UI component draws from a single source of truth. Without this package, components would hard-code hex values and drift from the design spec.

**Acceptance criteria**
- [ ] `internal/theme/colors.go` exports all color constants from PRD §10 as typed `lipgloss.Color` values: `Ink0` through `Ink9`, `Ember500`, `Plasma500`, `Mint500`, `Crimson500`, `Solar500`, `SynKeyword`, `SynString`, `SynFlag`
- [ ] `internal/theme/styles.go` exports pre-built `lipgloss.Style` values: `BlockBorderDefault`, `BlockBorderFocused`, `HeaderBg`, `InputBg`, `CanvasBg`, `AICardBg`, `MutedText`, `BodyText`, `BadgeSuccess`, `BadgeFailed`, `BadgeRunning`
- [ ] `BlockBorderFocused` uses ember-500 (#ff6b3d) as the border foreground color with a rounded border
- [ ] `BlockBorderDefault` uses ink-3 (#1e2230) as the border foreground color
- [ ] Badge styles use padding `(0, 1)` and appropriate background colors from PRD §10
- [ ] Package has no imports other than `lipgloss`
- [ ] `go test ./internal/theme/...` passes with assertions on at least three color hex values

**Tech notes**
- PRD §10 is the sole source of truth for hex values — do not invent or approximate
- `lipgloss.Color` accepts a hex string: `lipgloss.Color("#ff6b3d")`
- Additional token from DESIGN.md Ink ramp needed: Ink3 (`#1e2230`) for default border, Ink7 (`#8b91a8`) for header secondary text
- `BadgeRunning` text should read "in progress" (PRD §5.1 block states table)
- Do not add any Bubble Tea imports — this package is styling only

**Completion notes**
*(empty)*
