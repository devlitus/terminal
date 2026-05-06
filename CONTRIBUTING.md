# Contributing to Forge

## Workflow — Trunk-based development (light)

Forge uses **trunk-based development**. `main` is the single source of truth and must always build.

### Rules

- **Commit directly to `main`** for changes that take less than a day.
- **Use a short-lived branch** (max 1–2 days) only for changes that require multiple commits before they are stable. Branch naming: `feat/<slug>`, `fix/<slug>`, `chore/<slug>`.
- **Never commit broken builds** to `main`. Run `go build ./...` locally before pushing.
- **Feature flags** over long-lived branches. If a feature is incomplete, gate it behind a flag — don't sit on a branch.

### Commit messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <short summary>

feat(blocks): add ring buffer for 500-block cap
fix(acp): handle agent connection timeout gracefully
chore(ci): add golangci-lint to GitHub Actions
refactor(agent): extract OpenAI client into own package
docs(prd): resolve open questions OQ-1 through OQ-3
```

Types: `feat` · `fix` · `refactor` · `chore` · `docs` · `test` · `perf`

Scopes (use these consistently): `blocks` · `acp` · `agent` · `input` · `palette` · `config` · `ci` · `tui`

### Releases

Tag `main` when a milestone is stable:

```
git tag v0.1.0
git push origin v0.1.0
```

Versioning follows [semver](https://semver.org/). Pre-1.0: `v0.MINOR.PATCH`.

---

## CI

GitHub Actions runs on every push to `main` and on every PR (if branches are used):

| Check | Command |
|-------|---------|
| Build | `go build ./...` |
| Lint | `golangci-lint run` |
| Tests | `go test ./...` |

A push that fails CI must be fixed before the next feature work starts.

---

## Code style

- `gofmt` — always. Run before committing or configure your editor to run on save.
- `golangci-lint` — the project lint config is in `.golangci.yml`.
- No `//nolint` comments without a comment explaining why.
- Error wrapping: use `fmt.Errorf("context: %w", err)` — never discard errors silently.

---

## For AI agents

When working on this codebase, agents must:

1. Read `PRD.md` before starting any task — it defines scope, requirements, and explicit out-of-scope items.
2. Read `DESIGN.md` before touching any UI layer — all colors, spacing, and component patterns are defined there.
3. Commit atomically. One logical change per commit.
4. Never add dependencies without a clear reason tied to a PRD requirement.
5. Run `go build ./...` mentally — do not write code that cannot compile.
6. Follow the package structure defined in `ARCHITECTURE.md` once it exists.
