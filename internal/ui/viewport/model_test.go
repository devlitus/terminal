package viewport

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	datablock "github.com/forge-tui/forge/internal/block"
	"github.com/forge-tui/forge/internal/messages"
)

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

	// k: 2 → 1
	m, _ = pressKey(m, "k")
	if m.FocusedIdx() != 1 {
		t.Errorf("after first k: want 1, got %d", m.FocusedIdx())
	}

	// k: 1 → 0
	m, _ = pressKey(m, "k")
	if m.FocusedIdx() != 0 {
		t.Errorf("after second k: want 0, got %d", m.FocusedIdx())
	}

	// k at 0: stays at 0
	m, _ = pressKey(m, "k")
	if m.FocusedIdx() != 0 {
		t.Errorf("k at boundary: want 0, got %d", m.FocusedIdx())
	}

	// j: 0 → 1
	m, _ = pressKey(m, "j")
	if m.FocusedIdx() != 1 {
		t.Errorf("after first j: want 1, got %d", m.FocusedIdx())
	}

	// j: 1 → 2
	m, _ = pressKey(m, "j")
	if m.FocusedIdx() != 2 {
		t.Errorf("after second j: want 2, got %d", m.FocusedIdx())
	}

	// j at 2: stays at 2
	m, _ = pressKey(m, "j")
	if m.FocusedIdx() != 2 {
		t.Errorf("j at boundary: want 2, got %d", m.FocusedIdx())
	}
}

func TestBlockFocusedMsgEmitted(t *testing.T) {
	// 3 blocks with IDs 1, 2, 3; autoScroll leaves focusedIdx at 2 (ID=3)
	m := makeModel(3)

	// k moves focus from index 2 (ID=3) → index 1 (ID=2)
	_, cmd := pressKey(m, "k")

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
		m, _ = pressKey(m, "k")

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
