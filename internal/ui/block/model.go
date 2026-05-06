package block

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	datablock "github.com/forge-tui/forge/internal/block"
	"github.com/forge-tui/forge/internal/theme"
)

// Model renders a single command block with header, input, and output sections.
type Model struct {
	block      datablock.Block
	focused    bool
	width      int
	termHeight int
}

func New(b datablock.Block) Model {
	return Model{block: b, termHeight: 24}
}

func (m *Model) SetWidth(n int)      { m.width = n }
func (m *Model) SetFocused(f bool)   { m.focused = f }
func (m *Model) SetTermHeight(h int) { m.termHeight = h }

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.termHeight = msg.Height
	}
	return m, nil
}

func (m Model) View() string {
	w := m.width
	if w == 0 {
		w = 80
	}

	content := strings.Join([]string{
		m.renderHeader(w),
		m.renderInput(w),
		m.renderOutput(w),
	}, "\n")

	if m.focused {
		return theme.BlockBorderFocused.Render(content)
	}
	return theme.BlockBorderDefault.Render(content)
}

func (m Model) renderHeader(w int) string {
	left := theme.BodyText.Render(m.block.Dir)
	right := m.renderHeaderRight()

	gap := w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	inner := left + strings.Repeat(" ", gap) + right
	return theme.HeaderBg.Width(w).Render(inner)
}

func (m Model) renderHeaderRight() string {
	badge := m.renderBadge()
	if m.block.State == datablock.StateRunning {
		return badge
	}
	dur := fmt.Sprintf("%.1fs", m.block.Duration.Seconds())
	return theme.MutedText.Render(dur) + " " + badge
}

func (m Model) renderBadge() string {
	switch m.block.State {
	case datablock.StateSuccess:
		return theme.BadgeSuccess.Render("success")
	case datablock.StateFailed:
		return theme.BadgeFailed.Render("failed")
	default:
		return theme.BadgeRunning.Render("running")
	}
}

func (m Model) renderInput(w int) string {
	prompt := lipgloss.NewStyle().Foreground(theme.Ember500).Render("❯")
	cmd := lipgloss.NewStyle().Foreground(theme.Ink9).Render(m.block.Command)
	return theme.InputBg.Width(w).Render(prompt + " " + cmd)
}

func (m Model) renderOutput(w int) string {
	cap := m.termHeight / 3
	if cap < 3 {
		cap = 3
	}

	lines := m.block.Output
	if len(lines) > cap {
		lines = lines[len(lines)-cap:]
	}

	text := strings.Join(lines, "\n")
	return theme.CanvasBg.Width(w).Render(
		lipgloss.NewStyle().Foreground(theme.Ink8).Render(text),
	)
}
