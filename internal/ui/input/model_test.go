package input

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/forge-tui/forge/internal/messages"
)

func TestSubmitShellCommand(t *testing.T) {
	m := New()
	m.textinput.SetValue("git status")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a cmd, got nil")
	}
	msg := cmd()
	got, ok := msg.(messages.SubmitMsg)
	if !ok {
		t.Fatalf("expected SubmitMsg, got %T", msg)
	}
	if got.Input != "git status" {
		t.Errorf("Input = %q, want %q", got.Input, "git status")
	}
	if got.IsAIPrompt {
		t.Error("IsAIPrompt = true, want false")
	}
}

func TestSubmitAIPrompt(t *testing.T) {
	m := New()
	m.textinput.SetValue("/fix my code")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a cmd, got nil")
	}
	msg := cmd()
	got, ok := msg.(messages.SubmitMsg)
	if !ok {
		t.Fatalf("expected SubmitMsg, got %T", msg)
	}
	if got.Input != "/fix my code" {
		t.Errorf("Input = %q, want %q", got.Input, "/fix my code")
	}
	if !got.IsAIPrompt {
		t.Error("IsAIPrompt = false, want true")
	}
}

func TestOpenPalette(t *testing.T) {
	m := New()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlK})
	if cmd == nil {
		t.Fatal("expected a cmd, got nil")
	}
	msg := cmd()
	if _, ok := msg.(messages.OpenPaletteMsg); !ok {
		t.Fatalf("expected OpenPaletteMsg, got %T", msg)
	}
}

func TestClearCommand(t *testing.T) {
	m := New()
	m.textinput.SetValue("clear")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a cmd, got nil")
	}
	msg := cmd()
	if _, ok := msg.(messages.ViewportClearMsg); !ok {
		t.Fatalf("expected ViewportClearMsg, got %T", msg)
	}
}
