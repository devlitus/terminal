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
}

func New() Model {
	ti := textinput.New()
	ti.Placeholder = "type a command or /prompt for AI…"
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(theme.Ink6)
	ti.TextStyle = lipgloss.NewStyle().Foreground(theme.Ink9)
	ti.Prompt = "❯ "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(theme.Ember500)
	ti.Focus()
	return Model{textinput: ti}
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
			if v == "clear" {
				return m, func() tea.Msg { return messages.ViewportClearMsg{} }
			}
			isAI := strings.HasPrefix(v, "/") || strings.HasPrefix(v, "@")
			return m, func() tea.Msg {
				return messages.SubmitMsg{Input: v, IsAIPrompt: isAI}
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
	return lipgloss.NewStyle().Background(theme.Ink1).Width(m.width).Render(m.textinput.View())
}
