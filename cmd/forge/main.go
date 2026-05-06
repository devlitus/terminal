package main

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/forge-tui/forge/internal/block"
	forgeExec "github.com/forge-tui/forge/internal/exec"
	"github.com/forge-tui/forge/internal/messages"
	"github.com/forge-tui/forge/internal/ui/header"
	"github.com/forge-tui/forge/internal/ui/input"
	viewportui "github.com/forge-tui/forge/internal/ui/viewport"
)

// program is set in main() after tea.NewProgram so that exec goroutines
// can call program.Send() for streaming output messages.
var program *tea.Program

type rootModel struct {
	header  header.Model
	input   input.Model
	vp      viewportui.Model
	rb      *block.RingBuffer
	cancels map[string]context.CancelFunc
	cwd     string
}

func newRootModel() rootModel {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	return rootModel{
		header:  header.New(cwd),
		input:   input.New(),
		vp:      viewportui.New(),
		rb:      block.NewRingBuffer(),
		cancels: make(map[string]context.CancelFunc),
		cwd:     cwd,
	}
}

func (m rootModel) Init() tea.Cmd {
	return tea.Batch(m.header.Init(), m.input.Init(), m.vp.Init())
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		raw, c := m.header.Update(msg)
		m.header = raw.(header.Model)
		cmds = append(cmds, c)
		raw2, c2 := m.input.Update(msg)
		m.input = raw2.(input.Model)
		cmds = append(cmds, c2)
		raw3, c3 := m.vp.Update(msg)
		m.vp = raw3.(viewportui.Model)
		cmds = append(cmds, c3)
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if len(m.cancels) > 0 {
				// Cancel all running blocks; append ^C marker to their output.
				for blockID, cancel := range m.cancels {
					program.Send(messages.ExecOutputMsg{BlockID: blockID, Data: []byte("^C")})
					cancel()
					delete(m.cancels, blockID)
				}
				return m, nil
			}
			return m, tea.Quit
		}

	case messages.SubmitMsg:
		if !msg.IsAIPrompt {
			b := block.Block{
				Command:   msg.Input,
				Dir:       m.cwd,
				State:     block.StateRunning,
				StartedAt: time.Now(),
			}
			b = m.rb.Add(b)
			m.vp.AppendBlock(b)
			blockID := fmt.Sprintf("%d", b.ID)
			ctx, cancel := context.WithCancel(context.Background())
			m.cancels[blockID] = cancel
			return m, forgeExec.Start(program, blockID, b.Command, b.Dir, ctx)
		}

	case messages.ExecOutputMsg:
		raw, c := m.vp.Update(msg)
		m.vp = raw.(viewportui.Model)
		return m, c

	case messages.ExecDoneMsg:
		if cancel, ok := m.cancels[msg.BlockID]; ok {
			cancel()
			delete(m.cancels, msg.BlockID)
		}
		raw, c := m.vp.Update(msg)
		m.vp = raw.(viewportui.Model)
		return m, c

	case messages.ViewportClearMsg:
		m.vp = viewportui.New()
		m.rb = block.NewRingBuffer()
		return m, nil

	case messages.CwdChangedMsg:
		m.cwd = msg.Cwd
		raw, c := m.header.Update(msg)
		m.header = raw.(header.Model)
		return m, c
	}

	// Delegate remaining messages to input and viewport.
	raw, c := m.input.Update(msg)
	m.input = raw.(input.Model)
	cmds = append(cmds, c)

	raw2, c2 := m.vp.Update(msg)
	m.vp = raw2.(viewportui.Model)
	cmds = append(cmds, c2)

	return m, tea.Batch(cmds...)
}

func (m rootModel) View() string {
	return m.header.View() + "\n" + m.vp.View() + "\n" + m.input.View()
}

func main() {
	m := newRootModel()
	program = tea.NewProgram(m, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
