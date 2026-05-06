package header

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/forge-tui/forge/internal/messages"
)

func TestCwdDisplay(t *testing.T) {
	m := New("/home/user")
	out := m.View()
	if !strings.Contains(out, "/home/user") {
		t.Errorf("expected cwd '/home/user' in output, got: %q", out)
	}
}

func TestCwdTruncation(t *testing.T) {
	m := New("/home/user/projects/forge")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 15})
	out := updated.(Model).View()
	if !strings.Contains(out, "…") {
		t.Errorf("expected truncation ellipsis in output, got: %q", out)
	}
	if !strings.Contains(out, "forge") {
		t.Errorf("expected basename 'forge' in output, got: %q", out)
	}
}

func TestCwdChangedMsg(t *testing.T) {
	m := New("/home/user")
	updated, _ := m.Update(messages.CwdChangedMsg{Cwd: "/tmp"})
	out := updated.(Model).View()
	if !strings.Contains(out, "/tmp") {
		t.Errorf("expected updated cwd '/tmp' in output, got: %q", out)
	}
}
