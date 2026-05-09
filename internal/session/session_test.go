package session

import (
	"path/filepath"
	"testing"

	"github.com/forge-tui/forge/internal/block"
	"github.com/forge-tui/forge/internal/messages"
)

func TestClearResetsRingBufferAndEmitsMsg(t *testing.T) {
	rb := block.NewRingBuffer()
	rb.Add(block.Block{Command: "ls"})
	s := &Session{Cwd: t.TempDir(), rb: rb}

	handled, cmd := s.Handle("clear")
	if !handled {
		t.Fatal("expected handled=true for 'clear'")
	}
	if rb.Len() != 0 {
		t.Errorf("ring buffer should be empty after clear, got len=%d", rb.Len())
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd for 'clear'")
	}
	got := cmd()
	if _, ok := got.(messages.ViewportClearMsg); !ok {
		t.Errorf("expected ViewportClearMsg, got %T", got)
	}
}

func TestCdUpdatesDir(t *testing.T) {
	target := t.TempDir()
	rb := block.NewRingBuffer()
	s := &Session{Cwd: t.TempDir(), rb: rb}

	handled, cmd := s.Handle("cd " + target)
	if !handled {
		t.Fatal("expected handled=true for 'cd <path>'")
	}
	want := filepath.Clean(target)
	if s.Cwd != want {
		t.Errorf("Cwd=%q, want %q", s.Cwd, want)
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd for cd with argument")
	}
	got := cmd()
	cwdMsg, ok := got.(messages.CwdChangedMsg)
	if !ok {
		t.Fatalf("expected CwdChangedMsg, got %T", got)
	}
	if cwdMsg.Cwd != want {
		t.Errorf("CwdChangedMsg.Cwd=%q, want %q", cwdMsg.Cwd, want)
	}
}

func TestCdNoArg(t *testing.T) {
	origCwd := t.TempDir()
	rb := block.NewRingBuffer()
	s := &Session{Cwd: origCwd, rb: rb}

	handled, cmd := s.Handle("cd")
	if !handled {
		t.Fatal("expected handled=true for bare 'cd'")
	}
	if cmd != nil {
		t.Error("expected nil cmd for cd with no argument")
	}
	if s.Cwd != origCwd {
		t.Errorf("Cwd should be unchanged: got %q, want %q", s.Cwd, origCwd)
	}
}

func TestUnknownPassthrough(t *testing.T) {
	rb := block.NewRingBuffer()
	s := &Session{Cwd: t.TempDir(), rb: rb}

	handled, cmd := s.Handle("ls -la")
	if handled {
		t.Error("expected handled=false for 'ls -la'")
	}
	if cmd != nil {
		t.Error("expected nil cmd for unhandled input")
	}
}
