package acp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
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

// mockAgent simulates an ACP agent over in-process pipes.
// It reads one JSON-RPC request and responds with stream tokens + a done result.
func newMockClient(t *testing.T, respond func(agentIn io.Reader, agentOut io.Writer)) *Client {
	t.Helper()
	// Pipe 1: client writes → agent reads
	agentReader, clientWriter := io.Pipe()
	// Pipe 2: agent writes → client reads
	clientReader, agentWriter := io.Pipe()

	go respond(agentReader, agentWriter)

	c := &Client{
		stdin:     clientWriter,
		stdout:    bufio.NewReader(clientReader),
		connected: true,
	}
	c.nextID.Store(1) // pretend session/create already used ID=1
	return c
}

func TestSendPromptStreamsTokens(t *testing.T) {
	respond := func(agentIn io.Reader, agentOut io.Writer) {
		// Read (and discard) the prompt/turn request.
		scanner := bufio.NewScanner(agentIn)
		scanner.Scan()

		// Respond with two stream tokens then a done result.
		reqID := int64(2) // nextID starts at 1 after Store(1), Add(1) gives 2
		fmt.Fprintf(agentOut, `{"jsonrpc":"2.0","method":"stream","params":{"token":"hello"}}`+"\n")
		fmt.Fprintf(agentOut, `{"jsonrpc":"2.0","method":"stream","params":{"token":" world"}}`+"\n")
		fmt.Fprintf(agentOut, `{"jsonrpc":"2.0","id":%d,"result":{"done":true}}`+"\n", reqID)
	}

	c := newMockClient(t, respond)
	var tokens []string
	err := c.SendPrompt(context.Background(), "test prompt", func(tok string) {
		tokens = append(tokens, tok)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.Join(tokens, "")
	if got != "hello world" {
		t.Errorf("expected 'hello world', got %q", got)
	}
}

func TestSendPromptContextCancellation(t *testing.T) {
	respond := func(agentIn io.Reader, agentOut io.Writer) {
		// Consume the request then hang — never respond.
		bufio.NewScanner(agentIn).Scan()
		// Block forever (pipe will be closed when test ends)
		time.Sleep(10 * time.Second)
	}

	c := newMockClient(t, respond)
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := c.SendPrompt(ctx, "test", func(string) {})
	elapsed := time.Since(start)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("SendPrompt took too long to cancel: %v", elapsed)
	}
}

func TestSendPromptRPCError(t *testing.T) {
	respond := func(agentIn io.Reader, agentOut io.Writer) {
		bufio.NewScanner(agentIn).Scan()
		fmt.Fprintf(agentOut, `{"jsonrpc":"2.0","id":2,"error":{"code":-32600,"message":"bad request"}}`+"\n")
	}

	c := newMockClient(t, respond)
	err := c.SendPrompt(context.Background(), "bad", func(string) {})
	if err == nil || !strings.Contains(err.Error(), "bad request") {
		t.Errorf("expected rpc error, got %v", err)
	}
}

