package exec

import (
	"context"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/forge-tui/forge/internal/messages"
)

// Start returns a tea.Cmd that runs command in dir under ctx.
// Intermediate output lines are dispatched to the program via p.Send(ExecOutputMsg).
// The returned tea.Cmd's final message is ExecDoneMsg.
func Start(p *tea.Program, blockID, command, dir string, ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		out := make(chan string, 64)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for line := range out {
				p.Send(messages.ExecOutputMsg{
					BlockID: blockID,
					Data:    []byte(line),
				})
			}
		}()

		exitCode, duration, _ := Run(ctx, dir, command, out)
		wg.Wait() // ensure all output is sent before signalling done

		return messages.ExecDoneMsg{
			BlockID:  blockID,
			ExitCode: exitCode,
			Duration: duration,
		}
	}
}
