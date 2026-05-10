package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds the application's agent configuration.
type Config struct {
	AgentAPIBase string `toml:"api_base"`
	AgentAPIKey  string `toml:"api_key"`
	AgentModel   string `toml:"model"`
}

// fileEnvelope maps the [agent] TOML section to Config.
type fileEnvelope struct {
	Agent Config `toml:"agent"`
}

// IsShellOnly reports whether agent integration is disabled (no API base configured).
func (c *Config) IsShellOnly() bool {
	return c.AgentAPIBase == ""
}

// Load reads ~/.config/forge/config.toml. A missing file is not an error;
// defaults are returned instead.
func Load() (*Config, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return defaults(), nil
	}
	return load(filepath.Join(base, "forge", "config.toml"))
}

// load reads a TOML config file at path, merging values on top of defaults.
// A missing file returns defaults without error.
func load(path string) (*Config, error) {
	env := fileEnvelope{Agent: *defaults()}
	_, err := toml.DecodeFile(path, &env)
	if err != nil {
		if os.IsNotExist(err) {
			return defaults(), nil
		}
		return nil, err
	}
	result := env.Agent
	return &result, nil
}

func defaults() *Config {
	return &Config{
		AgentAPIBase: "http://localhost:11434/v1",
		AgentAPIKey:  "",
		AgentModel:   "qwen3.6",
	}
}
