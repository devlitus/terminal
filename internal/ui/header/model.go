package header

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/forge-tui/forge/internal/messages"
	"github.com/forge-tui/forge/internal/theme"
)

// Model renders a single-line header bar showing the working directory on the
// left and the block count on the right.
type Model struct {
	cwd        string
	blockCount int
	width      int
	statusHint string
}

func New(cwd string) Model {
	return Model{cwd: cwd}
}

// SetStatusHint sets a secondary hint line shown below the main header in ink-6.
// Pass an empty string to clear it.
func (m *Model) SetStatusHint(hint string) { m.statusHint = hint }

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.CwdChangedMsg:
		m.cwd = msg.Cwd
	case tea.WindowSizeMsg:
		m.width = msg.Width
	}
	return m, nil
}

func (m Model) View() string {
	right := theme.MutedText.Render(fmt.Sprintf("%d blocks", m.blockCount))
	rightLen := lipgloss.Width(right)

	if m.width == 0 {
		return theme.BodyText.Render(m.cwd) + " " + right
	}

	availLeft := m.width - rightLen - 1
	if availLeft < 0 {
		availLeft = 0
	}

	leftText := truncateCwd(m.cwd, availLeft)
	left := theme.BodyText.Render(leftText)

	leftLen := lipgloss.Width(left)
	gap := m.width - leftLen - rightLen
	if gap < 1 {
		gap = 1
	}

	content := left + strings.Repeat(" ", gap) + right
	mainLine := lipgloss.NewStyle().Width(m.width).Render(content)
	if m.statusHint != "" {
		hint := lipgloss.NewStyle().Foreground(lipgloss.Color("#5a6178")).Render(m.statusHint)
		return mainLine + "\n" + hint
	}
	return mainLine
}

// truncateCwd shortens cwd to fit within maxWidth visual characters.
// It removes path components from the left, prefixing with "…/", but
// never removes the last component (basename).
func truncateCwd(cwd string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if lipgloss.Width(cwd) <= maxWidth {
		return cwd
	}

	// Normalize to forward slashes; split into non-empty components.
	normalized := filepath.ToSlash(cwd)
	parts := strings.Split(normalized, "/")
	var components []string
	for _, p := range parts {
		if p != "" {
			components = append(components, p)
		}
	}
	if len(components) == 0 {
		return cwd
	}

	// Remove components from the left until the candidate fits,
	// but always keep at least the basename (last component).
	for i := 0; i < len(components)-1; i++ {
		remaining := components[i+1:]
		candidate := "…/" + strings.Join(remaining, "/")
		if lipgloss.Width(candidate) <= maxWidth {
			return candidate
		}
	}

	// Cannot fit even with just the basename — return it anyway.
	return "…/" + components[len(components)-1]
}
