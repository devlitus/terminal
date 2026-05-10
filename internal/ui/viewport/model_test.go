package viewport

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	datablock "github.com/forge-tui/forge/internal/block"
	"github.com/forge-tui/forge/internal/messages"
)

// extractSubmitMsg recursively unwraps a tea.Cmd looking for a SubmitMsg.
func extractSubmitMsg(cmd tea.Cmd) (messages.SubmitMsg, bool) {
	if cmd == nil {
		return messages.SubmitMsg{}, false
	}
	switch msg := cmd().(type) {
	case messages.SubmitMsg:
		return msg, true
	case tea.BatchMsg:
		for _, c := range msg {
			if sm, ok := extractSubmitMsg(c); ok {
				return sm, true
			}
		}
	}
	return messages.SubmitMsg{}, false
}

// makeModel creates a Model pre-loaded with n blocks (IDs 1..n, commands "cmd1".."cmdn")
// and initialised with a 120×30 terminal size.
func makeModel(n int) Model {
	m := New()
	raw, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	m = raw.(Model)
	for i := 1; i <= n; i++ {
		m.AppendBlock(datablock.Block{
			ID:      uint64(i),
			Command: fmt.Sprintf("cmd%d", i),
			State:   datablock.StateSuccess,
		})
	}
	return m
}

// pressKey sends a single rune key message and returns the updated model and cmd.
func pressKey(m Model, key string) (Model, tea.Cmd) {
	raw, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return raw.(Model), cmd
}

// extractBlockFocusedMsg recursively unwraps a tea.Cmd (including BatchMsg)
// looking for a BlockFocusedMsg.
func extractBlockFocusedMsg(cmd tea.Cmd) (messages.BlockFocusedMsg, bool) {
	if cmd == nil {
		return messages.BlockFocusedMsg{}, false
	}
	switch msg := cmd().(type) {
	case messages.BlockFocusedMsg:
		return msg, true
	case tea.BatchMsg:
		for _, c := range msg {
			if bfm, ok := extractBlockFocusedMsg(c); ok {
				return bfm, true
			}
		}
	}
	return messages.BlockFocusedMsg{}, false
}

func TestFocusNavigation(t *testing.T) {
	m := makeModel(3)

	if m.FocusedIdx() != 2 {
		t.Fatalf("initial focusedIdx: want 2, got %d", m.FocusedIdx())
	}

	// up: 2 → 1
	m, _ = pressKey(m, "up")
	if m.FocusedIdx() != 1 {
		t.Errorf("after first up: want 1, got %d", m.FocusedIdx())
	}

	// up: 1 → 0
	m, _ = pressKey(m, "up")
	if m.FocusedIdx() != 0 {
		t.Errorf("after second up: want 0, got %d", m.FocusedIdx())
	}

	// up at 0: stays at 0
	m, _ = pressKey(m, "up")
	if m.FocusedIdx() != 0 {
		t.Errorf("up at boundary: want 0, got %d", m.FocusedIdx())
	}

	// down: 0 → 1
	m, _ = pressKey(m, "down")
	if m.FocusedIdx() != 1 {
		t.Errorf("after first down: want 1, got %d", m.FocusedIdx())
	}

	// down: 1 → 2
	m, _ = pressKey(m, "down")
	if m.FocusedIdx() != 2 {
		t.Errorf("after second down: want 2, got %d", m.FocusedIdx())
	}

	// down at 2: stays at 2
	m, _ = pressKey(m, "down")
	if m.FocusedIdx() != 2 {
		t.Errorf("down at boundary: want 2, got %d", m.FocusedIdx())
	}
}

func TestBlockFocusedMsgEmitted(t *testing.T) {
	// 3 blocks with IDs 1, 2, 3; autoScroll leaves focusedIdx at 2 (ID=3)
	m := makeModel(3)

	// up moves focus from index 2 (ID=3) → index 1 (ID=2)
	_, cmd := pressKey(m, "up")

	bfm, ok := extractBlockFocusedMsg(cmd)
	if !ok {
		t.Fatal("expected BlockFocusedMsg to be emitted, got none")
	}
	if bfm.BlockID != "2" {
		t.Errorf("BlockID: want \"2\", got %q", bfm.BlockID)
	}
}

func TestAutoScroll(t *testing.T) {
	t.Run("disabled after navigating up", func(t *testing.T) {
		m := makeModel(2) // focusedIdx=1, autoScroll=true

		// navigate up → autoScroll=false, focusedIdx=0
		m, _ = pressKey(m, "up")

		m.AppendBlock(datablock.Block{ID: 3, Command: "cmd3", State: datablock.StateSuccess})

		// autoScroll is false, so focusedIdx must not jump to the new last block
		if m.FocusedIdx() != 0 {
			t.Errorf("want focusedIdx 0 after AppendBlock with autoScroll=false, got %d", m.FocusedIdx())
		}
	})

	t.Run("enabled by default", func(t *testing.T) {
		m := makeModel(2) // focusedIdx=1, autoScroll=true

		m.AppendBlock(datablock.Block{ID: 3, Command: "cmd3", State: datablock.StateSuccess})

		// autoScroll=true, so focusedIdx must follow the new last block
		if m.FocusedIdx() != 2 {
			t.Errorf("want focusedIdx 2 after AppendBlock with autoScroll=true, got %d", m.FocusedIdx())
		}
	})
}

func TestCopyNoFocus(t *testing.T) {
	m := New() // focusedIdx == -1, no blocks
	// must not panic and must emit no SubmitMsg
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ctrl+y")})
	m = updated.(Model)
	if _, ok := extractSubmitMsg(cmd); ok {
		t.Error("y with no focus: unexpected SubmitMsg emitted")
	}
}

func TestRerunNoFocus(t *testing.T) {
	m := New() // focusedIdx == -1, no blocks
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ctrl+r")})
	m = updated.(Model)
	if _, ok := extractSubmitMsg(cmd); ok {
		t.Error("r with no focus: unexpected SubmitMsg emitted")
	}
}

func TestRerunRunningBlock(t *testing.T) {
	m := New()
	raw, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	m = raw.(Model)
	m.AppendBlock(datablock.Block{
		ID:      1,
		Command: "long-running",
		State:   datablock.StateRunning,
	})
	// focusedIdx == 0 (autoScroll), block is Running
	_, cmd := pressKey(m, "ctrl+r")
	if _, ok := extractSubmitMsg(cmd); ok {
		t.Error("r on StateRunning block: unexpected SubmitMsg emitted")
	}
}

func TestRerunIdleBlock(t *testing.T) {
	m := New()
	raw, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	m = raw.(Model)
	m.AppendBlock(datablock.Block{
		ID:      1,
		Command: "echo hello",
		State:   datablock.StateSuccess,
	})
	// focusedIdx == 0
	_, cmd := pressKey(m, "ctrl+r")
	sm, ok := extractSubmitMsg(cmd)
	if !ok {
		t.Fatal("r on StateSuccess block: expected SubmitMsg, got none")
	}
	if sm.Input != "echo hello" {
		t.Errorf("SubmitMsg.Input: want %q, got %q", "echo hello", sm.Input)
	}
	if sm.IsAIPrompt {
		t.Error("SubmitMsg.IsAIPrompt: want false, got true")
	}
}

func TestCopyGracefulOnUnavailable(t *testing.T) {
	// This test verifies that y on a block with output does not panic,
	// regardless of whether a clipboard is available in the test environment.
	m := New()
	raw, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	m = raw.(Model)
	m.AppendBlock(datablock.Block{
		ID:      1,
		Command: "echo hi",
		State:   datablock.StateSuccess,
		Output:  []string{"hello"},
	})
	// Must not panic; clipboard may or may not be available in CI.
	updated, _ := pressKey(m, "ctrl+y")
	_ = updated // statusNotice is either "" or "clipboard unavailable" — both are valid
}
