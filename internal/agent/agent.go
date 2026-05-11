// Package agent implements an agentic loop on top of the ACP client.
// It maintains conversation history and executes tool calls requested by the model.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/forge-tui/forge/internal/acp"
	"github.com/forge-tui/forge/internal/messages"
)

// Agent wraps an ACP client with conversation history and tool execution.
type Agent struct {
	client  *acp.Client
	tools   []Tool
	history []acp.Message
	mu      sync.Mutex
}

// New creates an Agent. cwd is a function that returns the current working
// directory at call time (it may change between tool calls as the user navigates).
// confirmFn is called by run_command before executing; returning false cancels it.
func New(client *acp.Client, cwd func() string, confirmFn func(command string) bool) *Agent {
	return &Agent{
		client: client,
		tools:  defaultTools(cwd, confirmFn),
	}
}

// Reset clears the conversation history.
func (a *Agent) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.history = nil
}

// Send returns a tea.Cmd that runs the agentic loop in a goroutine.
// blockID identifies the AI card in the viewport that receives streamed tokens.
// userInput is the raw user message.
func (a *Agent) Send(p *tea.Program, blockID, userInput string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		a.mu.Lock()
		// Append system prompt on first turn.
		if len(a.history) == 0 {
			a.history = append(a.history, acp.Message{
				Role:    "system",
				Content: systemPrompt,
			})
		}
		a.history = append(a.history, acp.Message{
			Role:    "user",
			Content: userInput,
		})
		a.mu.Unlock()

		// Build tool definitions list.
		toolDefs := make([]acp.ToolDef, len(a.tools))
		toolMap := make(map[string]Tool, len(a.tools))
		for i, t := range a.tools {
			toolDefs[i] = t.Def
			toolMap[t.Def.Function.Name] = t
		}

		// Agentic loop: call model → execute tools → repeat until no tool calls.
		for {
			a.mu.Lock()
			history := make([]acp.Message, len(a.history))
			copy(history, a.history)
			a.mu.Unlock()

			var content strings.Builder
			toolCalls, err := a.client.Chat(ctx, history, toolDefs, func(token string) {
				content.WriteString(token)
				p.Send(messages.ACPTokenMsg{BlockID: blockID, Token: token})
			})
			if err != nil {
				return messages.ACPDoneMsg{BlockID: blockID, Err: err}
			}

			// No tool calls → model replied with content, done.
			if len(toolCalls) == 0 {
				a.mu.Lock()
				a.history = append(a.history, acp.Message{
					Role:    "assistant",
					Content: content.String(),
				})
				a.mu.Unlock()
				return messages.ACPDoneMsg{BlockID: blockID}
			}

			// Append assistant tool-call message to history.
			a.mu.Lock()
			a.history = append(a.history, acp.Message{
				Role:      "assistant",
				ToolCalls: toolCalls,
			})
			a.mu.Unlock()

			// Execute each tool call and collect results.
			for _, tc := range toolCalls {
				tool, ok := toolMap[tc.Function.Name]

				// Notify the UI that a tool is being called.
				p.Send(messages.ACPTokenMsg{
					BlockID: blockID,
					Token:   fmt.Sprintf("\n⚙ %s…\n", tc.Function.Name),
				})

				var result string
				if !ok {
					result = fmt.Sprintf("unknown tool: %s", tc.Function.Name)
				} else {
					var execErr error
					result, execErr = tool.Execute(ctx, json.RawMessage(tc.Function.Arguments))
					if execErr != nil {
						result = fmt.Sprintf("error: %s", execErr.Error())
					}
				}

				// Append tool result to history.
				a.mu.Lock()
				a.history = append(a.history, acp.Message{
					Role:       "tool",
					ToolCallID: tc.ID,
					Name:       tc.Function.Name,
					Content:    result,
				})
				a.mu.Unlock()
			}
			// Continue the loop: model will see the tool results and respond.
		}
	}
}
