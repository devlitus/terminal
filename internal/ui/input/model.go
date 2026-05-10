package input

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/forge-tui/forge/internal/messages"
	"github.com/forge-tui/forge/internal/theme"
)

// Model is the input bar component. It wraps a bubbles textinput and
// emits SubmitMsg, OpenPaletteMsg, or ViewportClearMsg on the relevant keys.
type Model struct {
	textinput textinput.Model
	width     int
	shellMode bool
}

func New() Model {
	ti := textinput.New()
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(theme.Ink6)
	ti.TextStyle = lipgloss.NewStyle().Foreground(theme.Ink9)
	ti.Focus()
	m := Model{textinput: ti}
	m.SetShellMode(false)
	return m
}

// SetShellMode switches the input bar between chat mode (shellMode=false) and
// shell mode (shellMode=true), updating the placeholder, prompt symbol, and
// prompt color accordingly.
func (m *Model) SetShellMode(v bool) {
	m.shellMode = v
	if v {
		m.textinput.Placeholder = "type a command…"
		m.textinput.Prompt = "❯ "
		m.textinput.PromptStyle = lipgloss.NewStyle().Foreground(theme.Ember500)
	} else {
		m.textinput.Placeholder = "ask AI anything…"
		m.textinput.Prompt = "⬡ "
		m.textinput.PromptStyle = lipgloss.NewStyle().Foreground(theme.Plasma500)
	}
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			v := strings.TrimSpace(m.textinput.Value())
			if v == "" {
				return m, nil
			}
			m.textinput.SetValue("")
			if v == "clear" || strings.EqualFold(v, "cls") {
				return m, func() tea.Msg { return messages.ViewportClearMsg{} }
			}
			return m, func() tea.Msg {
				return messages.SubmitMsg{Input: v}
			}
		case tea.KeyCtrlK:
			return m, func() tea.Msg { return messages.OpenPaletteMsg{} }
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	}

	var cmd tea.Cmd
	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	w := m.width
	if w < 4 {
		w = 4
	}
	inner := w - 2
	content := m.textinput.View()
	padded := lipgloss.NewStyle().Width(inner).Render(content)
	return theme.BlockBorderInput.Render(padded)
}
