package aicard

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	datablock "github.com/forge-tui/forge/internal/block"
	"github.com/forge-tui/forge/internal/theme"
)

// Plasma300 is the token text color for AI suggestions.
const Plasma300 lipgloss.Color = "#c4b5ff"

var (
	bodyStyle = lipgloss.NewStyle().
			Foreground(Plasma300)

	contentStyle = lipgloss.NewStyle().
			Foreground(Plasma300)

	toolLineStyle = lipgloss.NewStyle().
			Foreground(theme.Ink6).
			Italic(true)

	leftBorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.ThickBorder()).
			BorderLeft(true).
			BorderForeground(theme.Plasma500).
			MarginLeft(1).
			MarginTop(1)
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
	}
	return m, nil
}

// View renders the AI card: a plasma-bordered body with token text,
// a blinking cursor while streaming, or action hints when done.
func (m Model) View() string {
	// bodyW accounts for: 1 char left border + 1 char left margin = 2 total overhead.
	bodyW := m.width - 2
	if bodyW < 1 {
		bodyW = 1
	}

	if m.card.ErrMsg != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ef4444"))
		body := bodyStyle.Width(bodyW).Render(errStyle.Render(m.card.ErrMsg))
		return leftBorderStyle.Render(body)
	}

	tokenText := strings.Join(m.card.Tokens, "")
	if m.card.Streaming {
		tokenText += "▋"
	}

	body := bodyStyle.Width(bodyW).Render(renderLines(tokenText))
	return leftBorderStyle.Render(body)
}

// renderLines applies per-line styling: tool status lines (⚙) get a muted
// italic style; all other lines get the normal plasma content color.
func renderLines(content string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "⚙") {
			lines[i] = toolLineStyle.Render(line)
		} else {
			lines[i] = contentStyle.Render(line)
		}
	}
	return strings.Join(lines, "\n")
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
