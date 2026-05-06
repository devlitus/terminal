---
id: T-21
status: backlog
agent: Code Expert
depends-on: [T-18]
branch: T-21/ci-readme
---

## T-21 · CI Finalization and README

**Context**
Completes the developer-facing infrastructure: a fully functional CI pipeline that enforces quality gates on every push, and a README that lets a new developer build and run Forge in under five minutes.

**Acceptance criteria**
- [ ] `.github/workflows/ci.yml` triggers on push to `main` and on all PRs; runs `go build ./...`, `golangci-lint run`, `go test -race ./...` — all must pass
- [ ] `.golangci.yml` enables: `gofmt`, `govet`, `errcheck`, `staticcheck`, `misspell`, `godot`
- [ ] CI is green on the current `main` branch before this task is closed
- [ ] `README.md` includes: one-paragraph description, prerequisites (Go 1.23, WSL on Windows), build steps (`go build ./cmd/forge -o forge`), run steps, and a link to `PRD.md`
- [ ] README includes a "Shell-only mode" section explaining how to run Forge without an AI agent config
- [ ] README does NOT promise features beyond PRD §13 success criteria
- [ ] `git tag v0.1.0` is applied to the final commit after CI is green

**Tech notes**
- CI scaffold from T-01 is the starting point — this task validates and extends it
- `golangci/golangci-lint-action@v6` in CI; `actions/setup-go@v5` with `go-version: '1.23'`
- `go vet ./...` and `go build ./...` must pass locally before the PR is opened
- Tag only after CI passes — do not pre-tag

**Completion notes**
*(empty)*
