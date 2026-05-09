package aicard

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	datablock "github.com/forge-tui/forge/internal/block"
	"github.com/forge-tui/forge/internal/messages"
	"github.com/forge-tui/forge/internal/theme"
)

// Plasma300 is the token text color for AI suggestions.
const Plasma300 lipgloss.Color = "#c4b5ff"

var (
	bodyStyle = lipgloss.NewStyle().
			Background(theme.Ink1).
			Foreground(Plasma300)

	leftBorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.ThickBorder()).
			BorderLeft(true).
			BorderForeground(theme.Plasma500)
)

// Model is the Bubble Tea component for an AI suggestion card.
type Model struct {
	card  datablock.AICard
	width int
}

// New creates a new aicard Model for the given AICard data.
func New(card datablock.AICard) Model {
	return Model{card: card}
}

// SetWidth sets the render width of the card.
func (m *Model) SetWidth(n int) { m.width = n }

// Init satisfies tea.Model. No startup commands are needed.
func (m Model) Init() tea.Cmd { return nil }

// Update handles key events and window resize for the card.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyMsg:
		if m.card.Streaming {
			break
		}
		switch msg.Type {
		case tea.KeyEnter:
			return m, func() tea.Msg {
				return messages.AcceptAIMsg{Command: lastNonEmptyLine(m.card.Tokens)}
			}
		case tea.KeyEsc:
			return m, func() tea.Msg { return messages.DismissAIMsg{} }
		case tea.KeyRunes:
			if len(msg.Runes) > 0 && msg.Runes[0] == 'd' {
				return m, func() tea.Msg { return messages.DismissAIMsg{} }
			}
		}
	}
	return m, nil
}

// View renders the AI card: a plasma-bordered body with token text,
// a blinking cursor while streaming, or action hints when done.
func (m Model) View() string {
	if m.card.ErrMsg != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ef4444"))
		body := bodyStyle.Render(errStyle.Render(m.card.ErrMsg))
		return leftBorderStyle.Render(body)
	}

	tokenText := strings.Join(m.card.Tokens, "")

	var content string
	if m.card.Streaming {
		content = tokenText + "▋"
	} else {
		hints := theme.MutedText.Render("[Enter] Run  [Esc] Dismiss")
		content = tokenText + "\n" + hints
	}

	body := bodyStyle.Render(content)
	return leftBorderStyle.Render(body)
}

// lastNonEmptyLine returns the last non-blank line from the joined tokens,
// trimmed of surrounding whitespace. Falls back to the full joined string.
func lastNonEmptyLine(tokens []string) string {
	joined := strings.Join(tokens, "")
	lines := strings.Split(joined, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			return strings.TrimSpace(lines[i])
		}
	}
	return joined
}
