---
id: T-01
status: backlog
agent: Code Expert
depends-on: []
branch: T-01/project-scaffold
---

## T-01 · Project Scaffold

**Context**
Establishes the Go module, package directory structure, tooling config, and CI skeleton that every subsequent task builds on. Nothing else can start until this exists.

**Acceptance criteria**
- [ ] `go.mod` initialised with module path `github.com/forge-tui/forge` and `go 1.23`
- [ ] Directory tree created: `cmd/forge/`, `internal/theme/`, `internal/block/`, `internal/exec/`, `internal/config/`, `internal/acp/`, `internal/ui/block/`, `internal/ui/header/`, `internal/ui/input/`, `internal/ui/aicard/`, `internal/ui/viewport/`, `internal/ui/palette/`, `internal/session/`
- [ ] `go.mod` and `go.sum` include `github.com/charmbracelet/bubbletea` (≥ v2), `github.com/charmbracelet/lipgloss` (≥ v1), `github.com/charmbracelet/bubbles`
- [ ] `.golangci.yml` present with `gofmt`, `govet`, `errcheck`, `staticcheck` linters enabled
- [ ] `.github/workflows/ci.yml` runs `go build ./...`, `golangci-lint run`, `go test ./...` on push to `main` and on all PRs
- [ ] `cmd/forge/main.go` contains a stub `main()` that compiles cleanly
- [ ] `go build ./...` passes with zero errors

**Tech notes**
- Module path: `github.com/forge-tui/forge` (must be consistent across all packages)
- Run `go mod tidy` after adding all dependencies in one pass
- CI: `actions/setup-go@v5` with `go-version: '1.23'`; linter: `golangci/golangci-lint-action@v6`
- Each `internal/` subdirectory needs at least a placeholder `.go` file with a `package` declaration so the build graph is valid

**Completion notes**
*(empty)*
