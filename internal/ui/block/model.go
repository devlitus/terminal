package block

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	datablock "github.com/forge-tui/forge/internal/block"
	"github.com/forge-tui/forge/internal/theme"
	aicard "github.com/forge-tui/forge/internal/ui/aicard"
)

// ansiEscape matches ANSI/VT100 escape sequences that manipulate cursor
// position, clear the screen, or otherwise affect terminal layout.
// These sequences must be stripped from command output before rendering
// inside a block, or they will corrupt the TUI display (e.g. `cls`).
var ansiEscape = regexp.MustCompile(`\x1b(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])`)

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

// Block returns the underlying data block (used by the viewport for copy/re-run).
func (m Model) Block() datablock.Block { return m.block }

// AppendOutput appends a line of output text to the block.
// ANSI escape sequences that manipulate the terminal (cursor moves, screen
// clears, etc.) are stripped so they cannot corrupt the TUI layout.
func (m *Model) AppendOutput(data []byte) {
	if len(data) == 0 {
		return
	}
	clean := ansiEscape.ReplaceAllString(string(data), "")
	m.block.Output = append(m.block.Output, clean)
}

// SetDone transitions the block to its terminal state after the command exits.
func (m *Model) SetDone(exitCode int, dur time.Duration) {
	m.block.ExitCode = exitCode
	m.block.Duration = dur
	if exitCode == 0 {
		m.block.State = datablock.StateSuccess
	} else {
		m.block.State = datablock.StateFailed
	}
}

// StartAIStreaming initialises an AICard on this block and marks it as streaming.
func (m *Model) StartAIStreaming() {
	m.block.AICard = &datablock.AICard{Streaming: true}
}

// AppendAIToken appends a token to the active AICard.
func (m *Model) AppendAIToken(token string) {
	if m.block.AICard != nil {
		m.block.AICard.Tokens = append(m.block.AICard.Tokens, token)
	}
}

// SetAIStreamDone marks the AICard as no longer streaming.
func (m *Model) SetAIStreamDone() {
	if m.block.AICard != nil {
		m.block.AICard.Streaming = false
	}
}

// DismissAICard marks the AICard as dismissed so it no longer renders.
func (m *Model) DismissAICard() {
	if m.block.AICard != nil {
		m.block.AICard.Dismissed = true
	}
}

// SetAIError marks the AICard as errored with msg and stops streaming.
func (m *Model) SetAIError(msg string) {
	if m.block.AICard != nil {
		m.block.AICard.ErrMsg = msg
		m.block.AICard.Streaming = false
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.termHeight = msg.Height
	case tea.KeyMsg:
		// When an AI card is active and done streaming, delegate key events to it.
		if m.block.AICard != nil && !m.block.AICard.Dismissed && !m.block.AICard.Streaming {
			ac := aicard.New(*m.block.AICard)
			_, cmd := ac.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func (m Model) View() string {
	w := m.width
	if w == 0 {
		w = 80
	}
	// The rounded border adds 1 char on each side; inner content must be 2 narrower.
	innerW := w - 2
	if innerW < 1 {
		innerW = 1
	}

	parts := []string{
		m.renderHeader(innerW),
		m.renderInput(innerW),
		m.renderOutput(innerW),
	}
	if m.block.AICard != nil && !m.block.AICard.Dismissed {
		ac := aicard.New(*m.block.AICard)
		ac.SetWidth(innerW)
		parts = append(parts, ac.View())
	}
	content := strings.Join(parts, "\n")

	if m.focused {
		return theme.BlockBorderFocused.Render(content)
	}
	return theme.BlockBorderDefault.Render(content)
}

func (m Model) renderHeader(w int) string {
	left := m.renderHeaderLeft()
	right := m.renderBadge()

	gap := w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	inner := left + strings.Repeat(" ", gap) + right
	return lipgloss.NewStyle().Width(w).Render(inner)
}

func (m Model) renderHeaderLeft() string {
	sep := theme.MutedText.Render(" · ")

	dotColor := theme.Solar500
	switch m.block.State {
	case datablock.StateSuccess:
		dotColor = theme.Mint500
	case datablock.StateFailed:
		dotColor = theme.Crimson500
	}
	dot := lipgloss.NewStyle().Foreground(dotColor).Render("●")

	home, _ := os.UserHomeDir()
	dir := strings.Replace(m.block.Dir, home, "~", 1)
	path := theme.BodyText.Render(dir)

	if m.block.State == datablock.StateRunning {
		return dot + " " + path + sep + theme.BodyText.Render("running")
	}

	exit := theme.BodyText.Render(fmt.Sprintf("exit %d", m.block.ExitCode))
	dur := theme.BodyText.Render(fmt.Sprintf("%.1fs", m.block.Duration.Seconds()))
	return dot + " " + path + sep + exit + sep + dur
}

func (m Model) renderBadge() string {
	switch m.block.State {
	case datablock.StateSuccess:
		return lipgloss.NewStyle().Foreground(theme.Ink0).Background(theme.Mint500).Padding(0, 1).Render("✓ success")
	case datablock.StateFailed:
		return lipgloss.NewStyle().Foreground(theme.Ink0).Background(theme.Crimson500).Padding(0, 1).Render("✗ failed")
	default:
		return lipgloss.NewStyle().Foreground(theme.Ink0).Background(theme.Solar500).Padding(0, 1).Render("running")
	}
}

func (m Model) renderInput(w int) string {
	prompt := lipgloss.NewStyle().Foreground(theme.Ember500).Render("❯")
	cmd := lipgloss.NewStyle().Foreground(theme.Ink9).Render(m.block.Command)
	return lipgloss.NewStyle().Width(w).Render(prompt + " " + cmd)
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
	return lipgloss.NewStyle().Width(w).Foreground(theme.Ink8).Render(text)
}
