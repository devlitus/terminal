package session

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/forge-tui/forge/internal/block"
	"github.com/forge-tui/forge/internal/messages"
)

var cdRe = regexp.MustCompile(`^cd(\s+(.+))?$`)

// Session tracks the working directory and owns the ring buffer.
// It intercepts built-in commands (clear, cd) before they reach the shell.
type Session struct {
	Cwd string
	rb  *block.RingBuffer
}

// New creates a Session whose Cwd is initialised from os.Getwd().
func New(rb *block.RingBuffer) *Session {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	return &Session{Cwd: cwd, rb: rb}
}

// Handle intercepts built-in commands.
// Returns (true, cmd) when the input was consumed; the caller must not shell-exec it.
// Returns (false, nil) when the input is not a built-in.
func (s *Session) Handle(input string) (bool, tea.Cmd) {
	input = strings.TrimSpace(input)

	if input == "clear" {
		s.rb.Reset()
		return true, func() tea.Msg { return messages.ViewportClearMsg{} }
	}

	if m := cdRe.FindStringSubmatch(input); m != nil {
		arg := strings.TrimSpace(m[2])
		if arg == "" {
			return true, nil // cd with no argument is a no-op
		}
		var newPath string
		if filepath.IsAbs(arg) {
			newPath = filepath.Clean(arg)
		} else {
			newPath = filepath.Join(s.Cwd, arg)
		}
		s.Cwd = newPath
		return true, func() tea.Msg { return messages.CwdChangedMsg{Cwd: newPath} }
	}

	return false, nil
}
