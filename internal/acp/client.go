package acp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	osExec "os/exec"
	"sync"
	"sync/atomic"

	"github.com/forge-tui/forge/internal/config"
)

// ErrNotConnected is returned by SendPrompt when the agent subprocess is not running.
var ErrNotConnected = errors.New("acp: not connected")

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	ID      int64  `json:"id,omitempty"`
	Params  any    `json:"params"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method,omitempty"`
	ID      int64           `json:"id,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type streamParams struct {
	Token string `json:"token"`
}

// Client manages the lifecycle of an ACP agent subprocess over stdio.
type Client struct {
	cmd       *osExec.Cmd
	stdin     io.WriteCloser
	stdout    *bufio.Reader
	nextID    atomic.Int64
	mu        sync.Mutex // protects connected
	connected bool
	waitDone  chan struct{} // closed when background cmd.Wait() returns
	waitErr   error         // result of cmd.Wait(), read after waitDone is closed
}

// Connect starts the agent subprocess defined by the FORGE_AGENT_CMD env var
// (falling back to "forge-agent") and sends the session/create handshake.
// If cfg.IsShellOnly() is true, Connect is a no-op.
func (c *Client) Connect(cfg *config.Config) error {
	if cfg.IsShellOnly() {
		return nil
	}

	agentCmd := os.Getenv("FORGE_AGENT_CMD")
	if agentCmd == "" {
		agentCmd = "forge-agent"
	}

	path, err := osExec.LookPath(agentCmd)
	if err != nil {
		return fmt.Errorf("acp: agent binary %q not found: %w", agentCmd, err)
	}

	c.cmd = osExec.Command(path)

	stdin, err := c.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("acp: stdin pipe: %w", err)
	}
	c.stdin = stdin

	stdout, err := c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("acp: stdout pipe: %w", err)
	}
	c.stdout = bufio.NewReader(stdout)

	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("acp: start agent: %w", err)
	}

	c.connected = true

	// Background goroutine: detect subprocess death and clear connected.
	c.waitDone = make(chan struct{})
	go func() {
		c.waitErr = c.cmd.Wait()
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()
		close(c.waitDone)
	}()

	if err := c.sendRequest(rpcRequest{
		JSONRPC: "2.0",
		Method:  "session/create",
		ID:      c.nextID.Add(1),
		Params:  struct{}{},
	}); err != nil {
		return fmt.Errorf("acp: session/create: %w", err)
	}

	return nil
}

// sendRequest marshals req as a single JSON line and writes it to the agent stdin.
func (c *Client) sendRequest(req rpcRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = c.stdin.Write(data)
	return err
}

// SendPrompt sends a prompt/turn request and calls onToken for each streamed token.
// Returns ErrNotConnected if the subprocess is not running.
// Respects ctx cancellation between line reads.
func (c *Client) SendPrompt(ctx context.Context, prompt string, onToken func(string)) error {
	c.mu.Lock()
	connected := c.connected
	c.mu.Unlock()
	if !connected {
		return ErrNotConnected
	}

	reqID := c.nextID.Add(1)
	if err := c.sendRequest(rpcRequest{
		JSONRPC: "2.0",
		Method:  "prompt/turn",
		ID:      reqID,
		Params:  map[string]string{"prompt": prompt},
	}); err != nil {
		return fmt.Errorf("acp: prompt/turn: %w", err)
	}

	type readResult struct {
		line string
		err  error
	}

	for {
		// Check context before issuing the next blocking read.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Run the blocking read in a goroutine so context cancellation
		// can interrupt it. The goroutine is intentionally leaked only when
		// ctx is cancelled — it will exit as soon as the agent writes a line
		// or closes stdout.
		ch := make(chan readResult, 1)
		go func() {
			line, err := c.stdout.ReadString('\n')
			ch <- readResult{line, err}
		}()

		var res readResult
		select {
		case <-ctx.Done():
			return ctx.Err()
		case res = <-ch:
		}

		if res.err != nil {
			return fmt.Errorf("acp: read: %w", res.err)
		}

		var resp rpcResponse
		if err := json.Unmarshal([]byte(res.line), &resp); err != nil {
			return fmt.Errorf("acp: unmarshal: %w", err)
		}

		if resp.Error != nil {
			return fmt.Errorf("acp: rpc error %d: %s", resp.Error.Code, resp.Error.Message)
		}

		if resp.Method == "stream" {
			var sp streamParams
			if err := json.Unmarshal(resp.Params, &sp); err != nil {
				return fmt.Errorf("acp: stream params: %w", err)
			}
			onToken(sp.Token)
			continue
		}

		if resp.ID == reqID {
			break
		}
	}

	return nil
}

// Close sends session/close and waits for the subprocess to exit.
// If not connected, Close is a no-op.
func (c *Client) Close() error {
	c.mu.Lock()
	connected := c.connected
	c.mu.Unlock()
	if !connected {
		return nil
	}

	// Best-effort: ignore send error since we're shutting down.
	_ = c.sendRequest(rpcRequest{
		JSONRPC: "2.0",
		Method:  "session/close",
		ID:      c.nextID.Add(1),
		Params:  struct{}{},
	})

	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()
	_ = c.stdin.Close()

	// Wait for the background goroutine's cmd.Wait() instead of calling
	// Wait() a second time (double-Wait races and panics on some platforms).
	if c.waitDone != nil {
		<-c.waitDone
		return c.waitErr
	}
	return nil
}
