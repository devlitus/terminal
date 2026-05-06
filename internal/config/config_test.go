package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFile(t *testing.T) {
	cfg, err := load(filepath.Join(t.TempDir(), "nonexistent.toml"))
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	d := defaults()
	if cfg.AgentAPIBase != d.AgentAPIBase {
		t.Errorf("AgentAPIBase: got %q, want %q", cfg.AgentAPIBase, d.AgentAPIBase)
	}
	if cfg.AgentModel != d.AgentModel {
		t.Errorf("AgentModel: got %q, want %q", cfg.AgentModel, d.AgentModel)
	}
	if cfg.AgentAPIKey != d.AgentAPIKey {
		t.Errorf("AgentAPIKey: got %q, want %q", cfg.AgentAPIKey, d.AgentAPIKey)
	}
}

func TestLoadPartialTOML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	content := "[agent]\nmodel = \"llama3\"\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AgentModel != "llama3" {
		t.Errorf("AgentModel: got %q, want %q", cfg.AgentModel, "llama3")
	}
	d := defaults()
	if cfg.AgentAPIBase != d.AgentAPIBase {
		t.Errorf("AgentAPIBase: got %q, want %q", cfg.AgentAPIBase, d.AgentAPIBase)
	}
	if cfg.AgentAPIKey != d.AgentAPIKey {
		t.Errorf("AgentAPIKey: got %q, want %q", cfg.AgentAPIKey, d.AgentAPIKey)
	}
}

func TestIsShellOnly(t *testing.T) {
	if (&Config{AgentAPIBase: ""}).IsShellOnly() != true {
		t.Error("expected IsShellOnly() == true when AgentAPIBase is empty")
	}
	if (&Config{AgentAPIBase: "http://localhost"}).IsShellOnly() != false {
		t.Error("expected IsShellOnly() == false when AgentAPIBase is non-empty")
	}
}
