package acp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/forge-tui/forge/internal/config"
)

// ErrNotConnected is returned by SendPrompt when the agent is not connected.
var ErrNotConnected = errors.New("acp: not connected")

// --- Legacy ACP JSON-RPC types (kept for test compatibility via in-process pipes) ---

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

// Client manages AI agent communication.
//
// Production mode: calls the Ollama OpenAI-compatible HTTP API directly.
// Test mode: if stdin/stdout pipes are injected, speaks ACP JSON-RPC over them.
type Client struct {
	// HTTP mode fields
	apiBase    string
	apiKey     string
	model      string
	httpClient *http.Client

	// Legacy ACP/stdio fields — used only by tests that inject pipes directly.
	stdin    io.WriteCloser
	stdout   *bufio.Reader
	nextID   atomic.Int64
	mu       sync.Mutex
	connected bool
	waitDone  chan struct{}
	waitErr   error
}

// Connect initialises the client for the given config.
// In HTTP mode it stores the API parameters; no subprocess is launched.
// If cfg.IsShellOnly() is true, Connect is a no-op.
func (c *Client) Connect(cfg *config.Config) error {
	if cfg.IsShellOnly() {
		return nil
	}

	// If stdin was already injected (test mode), just mark as connected.
	if c.stdin != nil {
		c.mu.Lock()
		c.connected = true
		c.mu.Unlock()
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.apiBase = cfg.AgentAPIBase
	c.apiKey = cfg.AgentAPIKey
	c.model = cfg.AgentModel
	c.httpClient = &http.Client{}
	c.connected = true
	return nil
}

// SendPrompt sends a prompt and calls onToken for each streamed token.
// Returns ErrNotConnected if Connect has not been called successfully.
// Respects ctx cancellation.
func (c *Client) SendPrompt(ctx context.Context, prompt string, onToken func(string)) error {
	c.mu.Lock()
	connected := c.connected
	c.mu.Unlock()
	if !connected {
		return ErrNotConnected
	}

	// Use legacy ACP/stdio path when pipes are injected (test mode).
	if c.stdin != nil {
		return c.sendPromptACP(ctx, prompt, onToken)
	}

	return c.sendPromptHTTP(ctx, prompt, onToken)
}

// sendPromptHTTP calls the OpenAI-compatible /chat/completions endpoint with
// streaming and forwards each content token via onToken.
func (c *Client) sendPromptHTTP(ctx context.Context, prompt string, onToken func(string)) error {
	body := map[string]any{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"stream": true,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("acp: marshal request: %w", err)
	}

	url := strings.TrimRight(c.apiBase, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(data)))
	if err != nil {
		return fmt.Errorf("acp: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("acp: http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("acp: unexpected status %d", resp.StatusCode)
	}

	type sseChunk struct {
		Choices []struct {
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
		} `json:"choices"`
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")
		if payload == "[DONE]" {
			break
		}

		var chunk sseChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue // skip malformed chunks
		}
		if len(chunk.Choices) > 0 {
			if tok := chunk.Choices[0].Delta.Content; tok != "" {
				onToken(tok)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("acp: read stream: %w", err)
	}

	return nil
}

// sendPromptACP is the legacy JSON-RPC over stdio path, used by tests.
func (c *Client) sendPromptACP(ctx context.Context, prompt string, onToken func(string)) error {
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
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

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

// sendRequest marshals req as a single JSON line and writes it to stdin.
func (c *Client) sendRequest(req rpcRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = c.stdin.Write(data)
	return err
}

// Close cleans up the client. In HTTP mode this is a no-op.
// In ACP/stdio mode (tests) it closes the pipe and waits for any background goroutine.
func (c *Client) Close() error {
	c.mu.Lock()
	connected := c.connected
	c.mu.Unlock()
	if !connected {
		return nil
	}

	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()

	if c.stdin != nil {
		_ = c.stdin.Close()
		if c.waitDone != nil {
			<-c.waitDone
			return c.waitErr
		}
	}

	return nil
}
