package theme

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestColorHexValues(t *testing.T) {
	tests := []struct {
		name  string
		color lipgloss.Color
		want  string
	}{
		{"Ink0", Ink0, "#0a0b10"},
		{"Ink2", Ink2, "#161922"},
		{"Ink6", Ink6, "#5a6178"},
		{"Ink8", Ink8, "#c5c9d6"},
		{"Ember500", Ember500, "#ff6b3d"},
		{"Plasma500", Plasma500, "#7c5cff"},
		{"Mint500", Mint500, "#22c97e"},
		{"Crimson500", Crimson500, "#ef4444"},
		{"Solar500", Solar500, "#eab308"},
		{"SynKeyword", SynKeyword, "#7c5cff"},
		{"SynString", SynString, "#22c97e"},
		{"SynFlag", SynFlag, "#ff6b3d"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if string(tc.color) != tc.want {
				t.Errorf("color %s = %q, want %q", tc.name, tc.color, tc.want)
			}
		})
	}
}

// TestBorderStylesHaveBorder checks that border styles produce visible border
// characters regardless of terminal color support.
func TestBorderStylesHaveBorder(t *testing.T) {
	const roundedTopLeft = "╭"

	for _, tc := range []struct {
		name  string
		style lipgloss.Style
	}{
		{"BlockBorderDefault", BlockBorderDefault},
		{"BlockBorderFocused", BlockBorderFocused},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rendered := tc.style.Render("x")
			if !strings.Contains(rendered, roundedTopLeft) {
				t.Errorf("%s: expected rounded border, got %q", tc.name, rendered)
			}
		})
	}
}

// TestBadgesPadding checks that badge styles add horizontal padding.
func TestBadgesPadding(t *testing.T) {
	for _, tc := range []struct {
		name  string
		style lipgloss.Style
	}{
		{"BadgeSuccess", BadgeSuccess},
		{"BadgeFailed", BadgeFailed},
		{"BadgeRunning", BadgeRunning},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Padding(0,1) must surround the text with at least one space on each side.
			rendered := tc.style.Render("x")
			if !strings.Contains(rendered, " x ") {
				t.Errorf("%s: expected \" x \" with padding spaces, got %q", tc.name, rendered)
			}
		})
	}
}

// TestStylesRender verifies every exported style can be rendered without panic.
func TestStylesRender(t *testing.T) {
	styles := []struct {
		name  string
		style lipgloss.Style
	}{
		{"BlockBorderDefault", BlockBorderDefault},
		{"BlockBorderFocused", BlockBorderFocused},
		{"HeaderBg", HeaderBg},
		{"InputBg", InputBg},
		{"CanvasBg", CanvasBg},
		{"AICardBg", AICardBg},
		{"MutedText", MutedText},
		{"BodyText", BodyText},
		{"BadgeSuccess", BadgeSuccess},
		{"BadgeFailed", BadgeFailed},
		{"BadgeRunning", BadgeRunning},
	}

	for _, tc := range styles {
		t.Run(tc.name, func(t *testing.T) {
			// Must not panic and must produce non-empty output for non-empty input.
			if got := tc.style.Render("x"); got == "" {
				t.Errorf("%s: Render(\"x\") returned empty string", tc.name)
			}
		})
	}
}
