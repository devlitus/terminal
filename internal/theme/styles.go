package theme

import "github.com/charmbracelet/lipgloss"

// Block borders.
var (
	BlockBorderDefault = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Ink3)

	BlockBorderFocused = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Ember500)
)

// Surface backgrounds.
var (
	HeaderBg = lipgloss.NewStyle().Background(Ink2)
	InputBg  = lipgloss.NewStyle().Background(Ink1)
	CanvasBg = lipgloss.NewStyle().Background(Ink0)
	AICardBg = lipgloss.NewStyle().Background(Plasma500)
)

// Text styles.
var (
	MutedText = lipgloss.NewStyle().Foreground(Ink6)
	BodyText  = lipgloss.NewStyle().Foreground(Ink8)
)

// Status badges.
var (
	BadgeSuccess = lipgloss.NewStyle().
			Foreground(Ink0).
			Background(Mint500).
			Padding(0, 1)

	BadgeFailed = lipgloss.NewStyle().
			Foreground(Ink0).
			Background(Crimson500).
			Padding(0, 1)

	BadgeRunning = lipgloss.NewStyle().
			Foreground(Ink0).
			Background(Solar500).
			Padding(0, 1)
)
