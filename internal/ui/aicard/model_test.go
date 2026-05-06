package aicard_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	datablock "github.com/forge-tui/forge/internal/block"
	"github.com/forge-tui/forge/internal/messages"
	"github.com/forge-tui/forge/internal/ui/aicard"
)

func TestCursorDuringStreaming(t *testing.T) {
	card := datablock.AICard{Tokens: []string{"fix", " it"}, Streaming: true}
	m := aicard.New(card)
	view := m.View()
	if !strings.Contains(view, "▋") {
		t.Errorf("expected blinking cursor ▋ in streaming view, got: %q", view)
	}
}

func TestNoCursorWhenDone(t *testing.T) {
	card := datablock.AICard{Tokens: []string{"fix", " it"}, Streaming: false}
	m := aicard.New(card)
	view := m.View()
	if strings.Contains(view, "▋") {
		t.Errorf("expected no cursor ▋ when not streaming, got: %q", view)
	}
	if !strings.Contains(view, "[Enter] Run") {
		t.Errorf("expected action hints in done view, got: %q", view)
	}
}

func TestAcceptAIMsg(t *testing.T) {
	card := datablock.AICard{Tokens: []string{"echo hello"}, Streaming: false}
	m := aicard.New(card)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a command on Enter, got nil")
	}
	msg := cmd()
	accept, ok := msg.(messages.AcceptAIMsg)
	if !ok {
		t.Fatalf("expected AcceptAIMsg, got %T", msg)
	}
	if accept.Command != "echo hello" {
		t.Errorf("expected Command %q, got %q", "echo hello", accept.Command)
	}
}

func TestDismissAIMsg(t *testing.T) {
	card := datablock.AICard{Tokens: []string{"echo hello"}, Streaming: false}
	m := aicard.New(card)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected a command on Esc, got nil")
	}
	msg := cmd()
	if _, ok := msg.(messages.DismissAIMsg); !ok {
		t.Fatalf("expected DismissAIMsg, got %T", msg)
	}
}

func TestNoAcceptWhileStreaming(t *testing.T) {
	card := datablock.AICard{Tokens: []string{"echo hello"}, Streaming: true}
	m := aicard.New(card)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(messages.AcceptAIMsg); ok {
			t.Error("expected no AcceptAIMsg while streaming, but got one")
		}
	}
}
