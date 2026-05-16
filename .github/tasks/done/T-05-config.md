---
id: T-05
status: backlog
agent: Code Expert
depends-on: [T-01, T-02]
branch: T-05/config
---

## T-05 · Config Package

**Context**
Reads the optional `~/.config/forge/config.toml` and exposes a typed `Config` struct. The ACP client and root model use this to decide whether AI features are available at startup.

**Acceptance criteria**
- [ ] `internal/config/config.go` defines `Config` struct: `AgentAPIBase string`, `AgentAPIKey string`, `AgentModel string`
- [ ] `Load() (*Config, error)` reads `~/.config/forge/config.toml` if it exists; returns defaults when the file is absent — absent file is NOT an error
- [ ] Default values: `AgentAPIBase = "http://localhost:11434/v1"`, `AgentModel = "qwen2.5-coder:7b"`, `AgentAPIKey = ""`
- [ ] `(*Config).IsShellOnly() bool` returns `true` when `AgentAPIBase` is empty string
- [ ] TOML parsing uses `github.com/BurntSushi/toml` or `github.com/pelletier/go-toml/v2` — add to `go.mod`
- [ ] `go test ./internal/config/...` passes: missing file returns defaults, partial TOML merges with defaults, `IsShellOnly` returns correct value

**Tech notes**
- PRD §14 OQ-1 detail: config path is `~/.config/forge/config.toml`
- Use `os.UserConfigDir()` to resolve the config base directory portably (returns `$XDG_CONFIG_HOME` or `~/.config` on Linux)
- Do NOT read environment variables — out of scope for v1
- Do NOT create the config file if it is missing — Forge enters shell-only mode silently
- TOML section name must be `[agent]` to match the PRD example config

**Completion notes**
*(empty)*
