//go:build windows

package exec_test

import (
	"context"
	"testing"
	"time"

	"github.com/forge-tui/forge/internal/exec"
)

func TestRunExitZero(t *testing.T) {
	out := make(chan string, 16)
	code, _, err := exec.Run(context.Background(), t.TempDir(), "echo hello", out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	found := false
	for line := range out {
		if line == "hello" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected output to contain 'hello'")
	}
}

func TestRunExitNonZero(t *testing.T) {
	out := make(chan string, 16)
	code, _, err := exec.Run(context.Background(), t.TempDir(), "exit 1", out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
}

func TestRunCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	out := make(chan string, 16)

	done := make(chan struct{})
	go func() {
		defer close(done)
		exec.Run(ctx, ".", "ping -n 30 127.0.0.1", out) //nolint:errcheck
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(2000 * time.Millisecond):
		t.Fatal("Run did not return within 2000ms after context cancellation")
	}
}
