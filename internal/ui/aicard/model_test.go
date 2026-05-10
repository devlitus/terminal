package aicard_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	datablock "github.com/forge-tui/forge/internal/block"
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
}

func TestNoAcceptWhileStreaming(t *testing.T) {
	card := datablock.AICard{Tokens: []string{"echo hello"}, Streaming: true}
	m := aicard.New(card)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("expected no cmd while streaming on Enter")
	}
}
