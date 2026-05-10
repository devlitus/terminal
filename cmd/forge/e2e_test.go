package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/forge-tui/forge/internal/block"
	"github.com/forge-tui/forge/internal/messages"
)

// newTestModel returns a 120×40 root model for tests.
func newTestModel(t *testing.T) rootModel {
	t.Helper()
	m := newRootModel()
	raw, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	return raw.(rootModel)
}

// submitShell drives a shell SubmitMsg into the model and returns the model
// plus the blockID of the new block. Does NOT run the returned exec cmd.
func submitShell(m rootModel, cmd string) (rootModel, string) {
	raw, _ := m.Update(messages.SubmitMsg{Input: cmd})
	m = raw.(rootModel)
	all := m.rb.All()
	if len(all) == 0 {
		return m, ""
	}
	return m, fmt.Sprintf("%d", all[len(all)-1].ID)
}

// completeExec sends ExecOutputMsg lines then ExecDoneMsg to the model.
func completeExec(m rootModel, blockID string, exitCode int, lines ...string) rootModel {
	for _, line := range lines {
		raw, _ := m.Update(messages.ExecOutputMsg{BlockID: blockID, Data: []byte(line)})
		m = raw.(rootModel)
	}
	raw, _ := m.Update(messages.ExecDoneMsg{
		BlockID:  blockID,
		ExitCode: exitCode,
		Duration: 50 * time.Millisecond,
	})
	return raw.(rootModel)
}

// runBuiltin drives a builtin SubmitMsg, executes the returned cmd (unwrapping
// tea.BatchMsg if needed), and feeds each resulting message back into the model.
func runBuiltin(m rootModel, input string) rootModel {
	raw, cmd := m.Update(messages.SubmitMsg{Input: input})
	m = raw.(rootModel)
	if cmd == nil {
		return m
	}
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, sub := range batch {
			if sub != nil {
				raw2, _ := m.Update(sub())
				m = raw2.(rootModel)
			}
		}
		return m
	}
	raw2, _ := m.Update(msg)
	return raw2.(rootModel)
}

func TestShellOnlyModeHint(t *testing.T) {
	m := newTestModel(t)
	if !m.shellOnly {
		t.Skip("not in shellOnly mode")
	}
	if !strings.Contains(m.View(), "AI offline") {
		t.Errorf("expected View() to contain \"AI offline\"; got:\n%s", m.View())
	}
}

func TestSubmitCreatesRunningBlock(t *testing.T) {
	m := newTestModel(t)
	m, _ = submitShell(m, "echo hello")
	if m.rb.Len() != 1 {
		t.Fatalf("expected rb.Len() == 1, got %d", m.rb.Len())
	}
	b := m.rb.All()[0]
	if b.State != block.StateRunning {
		t.Errorf("expected StateRunning, got %s", b.State)
	}
	if b.Command != "echo hello" {
		t.Errorf("expected command \"echo hello\", got %q", b.Command)
	}
}

func TestExecDoneSuccess(t *testing.T) {
	m := newTestModel(t)
	m, blockID := submitShell(m, "echo hello")
	m = completeExec(m, blockID, 0)
	if got := m.vp.Blocks()[0].Block().State; got != block.StateSuccess {
		t.Errorf("expected StateSuccess, got %s", got)
	}
}

func TestExecDoneFailed(t *testing.T) {
	m := newTestModel(t)
	m, blockID := submitShell(m, "false")
	m = completeExec(m, blockID, 1)
	if got := m.vp.Blocks()[0].Block().State; got != block.StateFailed {
		t.Errorf("expected StateFailed, got %s", got)
	}
}

func TestBlockOutputAppearsInView(t *testing.T) {
	m := newTestModel(t)
	m, blockID := submitShell(m, "echo greetings")
	m = completeExec(m, blockID, 0, "greetings from forge")
	if !strings.Contains(m.View(), "greetings from forge") {
		t.Errorf("expected View() to contain \"greetings from forge\"")
	}
}

func TestCommandAppearsInView(t *testing.T) {
	m := newTestModel(t)
	m, blockID := submitShell(m, "ls -la")
	m = completeExec(m, blockID, 0)
	if !strings.Contains(m.View(), "ls -la") {
		t.Errorf("expected View() to contain \"ls -la\"")
	}
}

func TestMultipleBlocksAccumulate(t *testing.T) {
	m := newTestModel(t)
	for _, cmd := range []string{"cmd1", "cmd2", "cmd3"} {
		var blockID string
		m, blockID = submitShell(m, cmd)
		m = completeExec(m, blockID, 0)
	}
	if m.rb.Len() != 3 {
		t.Errorf("expected rb.Len() == 3, got %d", m.rb.Len())
	}
}

func TestClearViewport(t *testing.T) {
	m := newTestModel(t)
	var blockID string
	m, blockID = submitShell(m, "echo hi")
	m = completeExec(m, blockID, 0)
	m = runBuiltin(m, "clear")
	if m.rb.Len() != 0 {
		t.Errorf("expected rb.Len() == 0, got %d", m.rb.Len())
	}
	if len(m.vp.Blocks()) != 0 {
		t.Errorf("expected vp.Blocks() to be empty, got %d blocks", len(m.vp.Blocks()))
	}
}

func TestCommandPaletteOpensAndCloses(t *testing.T) {
	m := newTestModel(t)
	raw, _ := m.Update(messages.OpenPaletteMsg{})
	m = raw.(rootModel)
	if !m.paletteOpen {
		t.Error("expected paletteOpen == true after OpenPaletteMsg")
	}
	raw, _ = m.Update(messages.ClosePaletteMsg{})
	m = raw.(rootModel)
	if m.paletteOpen {
		t.Error("expected paletteOpen == false after ClosePaletteMsg")
	}
}

func TestQuitConfirmationWithRunningCommand(t *testing.T) {
	m := newTestModel(t)
	m, _ = submitShell(m, "sleep 10")
	raw, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ctrl+q")})
	m = raw.(rootModel)
	if !m.quitting {
		t.Fatal("expected quitting == true after pressing q with running command")
	}
	if !strings.Contains(m.View(), "Quit?") {
		t.Errorf("expected View() to contain \"Quit?\"; got:\n%s", m.View())
	}
	raw, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m = raw.(rootModel)
	if m.quitting {
		t.Error("expected quitting == false after pressing n")
	}
}

func TestCdChangesWorkingDirectory(t *testing.T) {
	m := newTestModel(t)
	m = runBuiltin(m, "cd /tmp")
	if m.session.Cwd != "/tmp" {
		t.Errorf("expected session.Cwd == \"/tmp\", got %q", m.session.Cwd)
	}
	if !strings.Contains(m.View(), "/tmp") {
		t.Errorf("expected View() to contain \"/tmp\"")
	}
}
