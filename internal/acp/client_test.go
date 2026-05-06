package acp

import (
	"context"
	"errors"
	"testing"
)

func TestSendPromptNotConnected(t *testing.T) {
	c := &Client{}
	err := c.SendPrompt(context.Background(), "hello", func(string) {})
	if !errors.Is(err, ErrNotConnected) {
		t.Fatalf("expected ErrNotConnected, got %v", err)
	}
}

func TestCloseNotConnected(t *testing.T) {
	c := &Client{}
	if err := c.Close(); err != nil {
		t.Fatalf("expected nil from Close on unconnected client, got %v", err)
	}
}
