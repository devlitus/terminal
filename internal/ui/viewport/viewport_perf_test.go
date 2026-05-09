package viewport

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	datablock "github.com/forge-tui/forge/internal/block"
)

// BenchmarkBlockViewportRender measures the cost of View() on a 500-block
// viewport at 120-column width.
func BenchmarkBlockViewportRender(b *testing.B) {
	m := makeModel(500)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = m.View()
	}
}

// TestViewportWidth80 verifies that rendering at 80 columns does not panic
// and produces non-empty output.
func TestViewportWidth80(t *testing.T) {
	m := New()
	raw, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	m = raw.(Model)

	m.AppendBlock(datablock.Block{ID: 1, Command: "ls", State: datablock.StateSuccess})

	view := m.View()
	if len(view) == 0 {
		t.Fatal("expected non-empty View() at width 80")
	}
}

// TestViewportWidth220 verifies that rendering at 220 columns does not panic
// and produces non-empty output.
func TestViewportWidth220(t *testing.T) {
	m := New()
	raw, _ := m.Update(tea.WindowSizeMsg{Width: 220, Height: 40})
	m = raw.(Model)

	m.AppendBlock(datablock.Block{ID: 1, Command: "ls", State: datablock.StateSuccess})

	view := m.View()
	if len(view) == 0 {
		t.Fatal("expected non-empty View() at width 220")
	}
}
