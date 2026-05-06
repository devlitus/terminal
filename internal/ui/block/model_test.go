package block_test

import (
	"strings"
	"testing"
	"time"

	datablock "github.com/forge-tui/forge/internal/block"
	"github.com/forge-tui/forge/internal/theme"
	uiblock "github.com/forge-tui/forge/internal/ui/block"
)

func newModel(state datablock.BlockState) uiblock.Model {
	b := datablock.Block{
		Command:  "echo hello",
		Dir:      "/home/user",
		State:    state,
		Duration: 1200 * time.Millisecond,
		Output:   []string{"hello"},
	}
	return uiblock.New(b)
}

func TestBadgeSuccess(t *testing.T) {
	m := newModel(datablock.StateSuccess)
	if !strings.Contains(m.View(), "success") {
		t.Error("expected 'success' badge in View() output")
	}
}

func TestBadgeFailed(t *testing.T) {
	m := newModel(datablock.StateFailed)
	if !strings.Contains(m.View(), "failed") {
		t.Error("expected 'failed' badge in View() output")
	}
}

func TestBadgeRunning(t *testing.T) {
	m := newModel(datablock.StateRunning)
	if !strings.Contains(m.View(), "running") {
		t.Error("expected 'running' badge in View() output")
	}
}

func TestFocusedBorderDiffersFromUnfocused(t *testing.T) {
	// Verify that focused and unfocused border styles are configured with
	// distinct colors. Full rendered output may look identical in no-color
	// environments (no ANSI codes emitted), so we compare style properties.
	fc := theme.BlockBorderFocused.GetBorderBottomForeground()
	dc := theme.BlockBorderDefault.GetBorderBottomForeground()
	if fc == dc {
		t.Errorf("border foreground colors should differ: focused=%v default=%v", fc, dc)
	}
}

func TestNoPanicAt80Cols(t *testing.T) {
	m := newModel(datablock.StateSuccess)
	m.SetWidth(80)
	_ = m.View()
}

func TestNoPanicAt220Cols(t *testing.T) {
	m := newModel(datablock.StateSuccess)
	m.SetWidth(220)
	_ = m.View()
}
