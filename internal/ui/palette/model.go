package palette

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"

	"github.com/forge-tui/forge/internal/messages"
	"github.com/forge-tui/forge/internal/theme"
)

var builtins = []string{"Re-run last command", "Copy last output", "Fix with AI"}

// Model is the Bubble Tea component for the command palette overlay.
type Model struct {
	commands []string // snapshot at open time, most-recent first (max 20)
	filtered []string // current filtered view of commands
	query    string
	cursor   int
	width    int
	height   int
	lastCmd  string // commands[0] if non-empty
}

// New creates a ready-to-use Model. commands must be most-recent-first, max 20.
func New(commands []string, width, height int) Model {
	last := ""
	if len(commands) > 0 {
		last = commands[0]
	}
	return Model{
		commands: commands,
		filtered: commands,
		width:    width,
		height:   height,
		lastCmd:  last,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return m, func() tea.Msg { return messages.ClosePaletteMsg{} }
		case tea.KeyEnter:
			return m, m.selectCurrent()
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown:
			if m.cursor < len(m.filtered)+len(builtins)-1 {
				m.cursor++
			}
		case tea.KeyBackspace:
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
				m.applyFilter()
				m.cursor = 0
			}
		case tea.KeyRunes:
			if len(msg.Runes) > 0 {
				m.query += string(msg.Runes)
				m.applyFilter()
				m.cursor = 0
			}
		}
	}
	return m, nil
}

func (m *Model) applyFilter() {
	if m.query == "" {
		m.filtered = m.commands
		return
	}
	matches := fuzzy.Find(m.query, m.commands)
	out := make([]string, 0, len(matches))
	for _, match := range matches {
		out = append(out, match.Str)
	}
	m.filtered = out
}

func (m Model) selectCurrent() tea.Cmd {
	if m.cursor < len(m.filtered) {
		cmd := m.filtered[m.cursor]
		return func() tea.Msg { return messages.SubmitMsg{Input: cmd, IsAIPrompt: false} }
	}

	switch m.cursor - len(m.filtered) {
	case 0: // Re-run last command
		if m.lastCmd == "" {
			return func() tea.Msg { return messages.ClosePaletteMsg{} }
		}
		return func() tea.Msg { return messages.SubmitMsg{Input: m.lastCmd, IsAIPrompt: false} }
	case 1: // Copy last output
		return func() tea.Msg { return messages.CopyLastOutputMsg{} }
	case 2: // Fix with AI
		return func() tea.Msg { return messages.FixWithAIMsg{} }
	}
	return nil
}

func (m Model) View() string {
	w := m.width - 4
	if w > 60 {
		w = 60
	}
	if w < 20 {
		w = 20
	}

	var sb strings.Builder
	sb.WriteString("  > " + m.query + "█\n\n")

	for i, item := range m.filtered {
		if i == m.cursor {
			sb.WriteString(lipgloss.NewStyle().Foreground(theme.Ember500).Render("> "+item) + "\n")
		} else {
			sb.WriteString("  " + item + "\n")
		}
	}

	if len(m.filtered) > 0 {
		sb.WriteString("\n")
	}

	for i, item := range builtins {
		idx := len(m.filtered) + i
		if idx == m.cursor {
			sb.WriteString(lipgloss.NewStyle().Foreground(theme.Plasma500).Render("> "+item) + "\n")
		} else {
			sb.WriteString(theme.MutedText.Render("  "+item) + "\n")
		}
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(theme.Ink8).
		Bold(true)

	boxStyle := lipgloss.NewStyle().
		Background(theme.Ink1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Ink5).
		Padding(1, 2).
		Width(w)

	body := titleStyle.Render("Command Palette") + "\n\n" + sb.String()
	return boxStyle.Render(body)
}
