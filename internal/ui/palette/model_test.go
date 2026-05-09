package palette

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/forge-tui/forge/internal/messages"
)

func TestFuzzyFilter(t *testing.T) {
	m := New([]string{"ls", "git status", "go test"}, 80, 24)

	for _, ch := range "go" {
		raw, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = raw.(Model)
	}

	found := false
	for _, f := range m.filtered {
		if f == "go test" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'go test' in filtered results, got %v", m.filtered)
	}
}

func TestEscClosePalette(t *testing.T) {
	m := New([]string{"ls"}, 80, 24)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}
	msg := cmd()
	if _, ok := msg.(messages.ClosePaletteMsg); !ok {
		t.Errorf("expected ClosePaletteMsg, got %T", msg)
	}
}

func TestBuiltinsAlwaysPresent(t *testing.T) {
	m := New([]string{}, 80, 24)
	view := m.View()
	for _, b := range builtins {
		if !strings.Contains(view, b) {
			t.Errorf("expected builtin %q in view", b)
		}
	}
}

func TestEnterSelectsCommand(t *testing.T) {
	m := New([]string{"ls", "git status"}, 80, 24)
	// cursor starts at 0; first filtered item is "ls"
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a cmd, got nil")
	}
	msg := cmd()
	sm, ok := msg.(messages.SubmitMsg)
	if !ok {
		t.Fatalf("expected SubmitMsg, got %T", msg)
	}
	if sm.Input != "ls" {
		t.Errorf("Input = %q, want %q", sm.Input, "ls")
	}
	if sm.IsAIPrompt {
		t.Error("IsAIPrompt = true, want false")
	}
}
